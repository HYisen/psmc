package main

import (
	"fmt"
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
		}
		fmt.Printf("%8d: %v\n", pid, commandName)
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
