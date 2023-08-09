package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// SMaps is the avatar of file /proc/pid/smaps or its
// ref https://man7.org/linux/man-pages/man5/proc.5.html
// The number unit is kB as I sampled in real system.
// All fields are the sum of each of the process's mappings.
// ref https://unix.stackexchange.com/questions/33381
type SMaps struct {
	PSS            int // Proportional Set Size
	Shared         int
	Private        int
	SharedHugeTLB  int // Translation LookAside Buffer
	PrivateHugeTLB int
}

// Absorb read the line and extract its value to corresponding field of SMaps.
// If not matched, would do nothing and return nil.
func (sm *SMaps) Absorb(line string) error {
	for _, oa := range orderedAbsorbers {
		if !strings.HasPrefix(line, oa.Prefix) {
			continue
		}
		num, err := numberInMappingLine(line)
		if err != nil {
			return err
		}
		oa.Handler(sm, num)
	}
	return nil
}

func numberInMappingLine(s string) (num int, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse mappingLine %s: %v", s, err)
		}
	}()

	_, rest, found := strings.Cut(s, ":")
	if !found {
		return 0, fmt.Errorf("no comma")
	}

	const unitSuffix = "kB"
	if !strings.HasSuffix(rest, unitSuffix) {
		return 0, fmt.Errorf("not ends with %s", unitSuffix)
	}
	numStr := strings.TrimSpace(strings.TrimSuffix(rest, unitSuffix))

	ret, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("parse numStr %s: %v", numStr, err)
	}
	return ret, nil
}

type Absorber struct {
	Prefix  string
	Handler func(sm *SMaps, num int)
}

// orderedAbsorbers are strategies of how to match and absorb mappingLine.
// MUST use in sequence order as some low-priority branches are super set of other high-priority branches.
// ref https://www.kernel.org/doc/Documentation/filesystems/proc.txt
// "Shared_Hugetlb" and "Private_Hugetlb" are not counted in "RSS" or "PSS" field for historical reasons.
// An ef-if blob is the original design, but its Circle Complexity is too high.
var orderedAbsorbers = []Absorber{
	{
		Prefix: "Pss:",
		Handler: func(sm *SMaps, num int) {
			sm.PSS += num
		},
	},
	{
		Prefix: "Shared_Hugetlb:",
		Handler: func(sm *SMaps, num int) {
			sm.SharedHugeTLB += num
		},
	},
	{
		Prefix: "Shared",
		Handler: func(sm *SMaps, num int) {
			sm.Shared += num
		},
	},
	{
		Prefix: "Private_Hugetlb:",
		Handler: func(sm *SMaps, num int) {
			sm.PrivateHugeTLB += num
		},
	},
	{
		Prefix: "Private",
		Handler: func(sm *SMaps, num int) {
			sm.Private += num
		},
	},
}

// NewSMaps reads the statm file by pid and return the parsed and summarized data.
func NewSMaps(pid int) (*SMaps, error) {
	fp, err := os.Open(path.Join(procPathRoot, strconv.Itoa(pid), "smaps_rollup"))
	if os.IsNotExist(err) {
		fp, err = os.Open(path.Join(procPathRoot, strconv.Itoa(pid), "smaps"))
	}
	if err != nil {
		return nil, err
	}

	var ret SMaps
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		if err := ret.Absorb(scanner.Text()); err != nil {
			return nil, err
		}
	}
	return &ret, nil
}
