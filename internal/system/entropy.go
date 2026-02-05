//
// internal/system/entropy.go
// 
package system

import (
	"os"
	"bufio"
	"strconv"	
)

func readEntropy() (uint64, error) {
	file, err := os.Open("/proc/sys/kernel/random/entropy_avail")

	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		val, err := strconv.ParseUint(line, 10, 64)

		if err != nil {
			continue
		}

		return val, nil
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return 0, err
}
