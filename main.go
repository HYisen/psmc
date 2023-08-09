package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	fmt.Println("Hello world.")
	pids, err := getPIDs()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pids)
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
