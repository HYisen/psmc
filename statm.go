package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// StatM is the avatar of file /proc/pid/statm.
// ref https://man7.org/linux/man-pages/man5/proc.5.html
type StatM struct {
	Size     int
	Resident int
	Shared   int
	Text     int
	Lib      int
	Data     int
	DT       int
}

// NewStatM reads the statm file by pid and return the parsed data.
func NewStatM(pid int) (*StatM, error) {
	bytes, err := os.ReadFile(path.Join(procPathRoot, strconv.Itoa(pid), "statm"))
	if err != nil {
		return nil, err
	}
	bytes = bytes[:len(bytes)-1] // remove tailing \n
	parts := strings.Split(string(bytes), " ")
	var nums []int
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("parse %v: %v", string(bytes), err)
		}
		nums = append(nums, num)
	}
	return &StatM{
		Size:     nums[0],
		Resident: nums[1],
		Shared:   nums[2],
		Text:     nums[3],
		Lib:      nums[4],
		Data:     nums[5],
		DT:       nums[6],
	}, nil
}
