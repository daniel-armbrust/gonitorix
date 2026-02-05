//
// internal/system/uptime.go
//
package system

import (
	"os"
	"bufio"
	"fmt"
	"strings"
	"strconv"
)

func uptimeToString(uptime float64) string {
	secs := int64(uptime)

	d := secs / (60 * 60 * 24)
	h := (secs / (60 * 60)) % 24
	m := (secs / 60) % 60

	var dStr string
	if d > 0 {
		dStr = fmt.Sprintf("%d days,", d)
	}

	var result string
	if h > 0 {
		result = fmt.Sprintf("%s %dh %dm", dStr, h, m)
	} else {
		result = fmt.Sprintf("%s %d min", dStr, m)
	}

	return strings.TrimSpace(result)
}

func readUptime() (float64, error) {
	file, err := os.Open("/proc/uptime")

	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	if scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)

		if len(fields) >= 1 {

			val, err := strconv.ParseFloat(fields[0], 64)

			if err == nil {
				return val, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return 0, fmt.Errorf("Uptime not found\n")
}