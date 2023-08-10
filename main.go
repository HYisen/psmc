package main

import (
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

func main() {
	pids, err := getPIDs()
	if err != nil {
		log.Fatal(err)
	}

	for _, pid := range pids {
		commandName, err := getCommandName(pid)
		if err != nil {
			continue
		}
		private, shared, err := getMemoryStat(pid)
		if err != nil {
			log.Println(err)
			continue
		}
		args, err := getArguments(pid)
		if err != nil {
			log.Println(err)
			continue
		}
		stat, err := NewStat(pid)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(
			"%6d(%6d): %v %s + %s\n",
			pid,
			stat.PPID,
			formatName(commandName, args),
			humanize.IBytes(uint64(private)),
			humanize.IBytes(uint64(shared)),
		)
	}
}

func formatName(commandName string, args []string) string {
	if len(args) > 0 && args[0] == commandName {
		return strings.Join(args, " ")
	}
	return fmt.Sprintf("%s %v", commandName, args)
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
// Or even as root, no such file or directory, as kernel threads don't have exe links or process gone.
func getCommandName(pid int) (string, error) {
	return os.Readlink(path.Join(procPathRoot, strconv.Itoa(pid), "exe"))
}

func getArguments(pid int) ([]string, error) {
	data, err := os.ReadFile(path.Join(procPathRoot, strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return nil, err
	}
	var ret []string
	// ref https://man7.org/linux/man-pages/man5/proc.5.html
	// The commandline arguments appear in this file as a set of strings separated by null bytes ('\0'),
	// with a further null byte after the last string.
	parts := bytes.Split(data[:len(data)-1], []byte{0})
	for _, part := range parts {
		ret = append(ret, string(part))
	}
	return ret, nil
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
