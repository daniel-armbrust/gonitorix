//
// internal/system/meminfo.go
// 
package system

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
)

func readMemory() (map[string]uint64, error) {
	file, err := os.Open("/proc/meminfo")

	if err != nil {
		return nil, err
	}
	defer file.Close()

	re := regexp.MustCompile(`^(\w+):\s+(\d+)\s+kB`)

	mem := make(map[string]uint64)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		m := re.FindStringSubmatch(line)

		if len(m) != 3 {
			continue
		}

		key := m[1]
		valStr := m[2]

		val, err := strconv.ParseUint(valStr, 10, 64)

		if err != nil {
			continue
		}

		switch key {
			case "MemTotal",
				 "MemFree",
				 "Buffers",
				 "Cached",
				 "Active",
				 "Inactive",
				 "SReclaimable",
				 "SUnreclaim":

				mem[key] = val

				if key == "SUnreclaim" {
					break
				}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// SReclaimable and SUnreclaim values are added to 'mfree'
	// in order to be also included in the subtraction later.
	if srecl, ok := mem["SReclaimable"]; ok {
		mem["MemFree"] += srecl
	}

	if sun, ok := mem["SUnreclaim"]; ok {
		mem["MemFree"] += sun
	}

	return mem, nil
}