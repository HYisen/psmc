package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"os"
	"path"
	"strconv"
)

func main() {
	fmt.Println("Hello world.")
	pids, err := getPIDs()
	if err != nil {
		log.Fatal(err)
	}

	for _, pid := range pids {
		commandName, err := getCommandName(pid)
		if err != nil {
			log.Println(err)
			continue
		}
		private, shared, err := getMemoryStat(pid)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Printf(
			"%8d: %v %s + %s\n",
			pid,
			commandName,
			humanize.IBytes(uint64(private)),
			humanize.IBytes(uint64(shared)),
		)
	}
}

const procPathRoot = "/proc"

func getPIDs() ([]int, error) {
	dirEntries, err := os.ReadDir(procPathRoot)
	if err != nil {
		return nil, err
	}

	var ret []int
	for _, entry := range dirEntries {
		num, err := strconv.Atoi(entry.Name())
		if err != nil {
			// some filenames are not uid numbers, skip in silence.
			continue
		}
		ret = append(ret, num)
	}
	return ret, nil
}

// getCommandName is getCmdName in ps_men.py, read the whole name of link /proc/${pid}/exe.
// Commonly return a permission denied error with path on other user's process if not run with root.
func getCommandName(pid int) (string, error) {
	return os.Readlink(path.Join(procPathRoot, strconv.Itoa(pid), "exe"))
}

var pageSize = os.Getpagesize()

func getMemoryStat(pid int) (private, shared int, err error) {
	smaps, err := NewSMaps(pid)
	if err != nil {
		if os.IsNotExist(err) {
			// fallback to statm calculation
			statm, err := NewStatM(pid)
			if err != nil {
				return 0, 0, err
			}
			private = (statm.Resident - statm.Shared) * pageSize
			shared = statm.Shared * pageSize
			return private, shared, nil
		}
		return 0, 0, err
	}

	if smaps.PSS != 0 {
		shared = (smaps.PSS - smaps.Private) * iecScale
	} else {
		shared = (smaps.Shared) * iecScale
	}
	private = (smaps.Private + smaps.PrivateHugeTLB) * iecScale
	return private, shared, nil
}

const iecScale = 1024
