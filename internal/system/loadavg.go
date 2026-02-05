//
// internal/system/loadavg.go
// 
package system

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
)

func readLoadAvg() (map[string]float64, error) {
	file, err := os.Open("/proc/loadavg")

	if err != nil {
		return nil, err
	}
	defer file.Close()

	re := regexp.MustCompile(`^(\d+\.\d+)\s+(\d+\.\d+)\s+(\d+\.\d+)`)

	load := make(map[string]float64)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		m := re.FindStringSubmatch(line)

		if len(m) == 4 {

			load1, _ := strconv.ParseFloat(m[1], 64)
			load5, _ := strconv.ParseFloat(m[2], 64)
			load15, _ := strconv.ParseFloat(m[3], 64)

			load["load1"] = load1
			load["load5"] = load5
			load["load15"] = load15

			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return load, nil
}