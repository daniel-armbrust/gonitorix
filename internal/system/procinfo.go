//
// internal/system/procinfo.go
// 
package system

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func readProcInfo() (map[string]uint64, error) {
	procstats := map[string]uint64{
		"run":    0,
		"sleep":  0,
		"wio":    0,
		"zombie": 0,
		"stop":   0,
		"swap":   0,
	}

	dirs, err := filepath.Glob("/proc/[0-9]*")

	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)

		if err != nil || !info.IsDir() {
			continue
		}

		statusFile := dir + "/status"

		if _, err := os.Stat(statusFile); err != nil {
			continue
		}

		f, err := os.Open(statusFile)

		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "State:") {

				fields := strings.Fields(line)

				if len(fields) >= 2 {

					state := fields[1]

					switch state {
					case "R":
						procstats["run"]++
					case "S":
						procstats["sleep"]++
					case "D":
						procstats["wio"]++
					case "Z":
						procstats["zombie"]++
					case "T":
						procstats["stop"]++
					case "W":
						procstats["swap"]++
					}
				}

				break
			}
		}

		f.Close()
	}

	procstats["total"] = procstats["run"]    +
						 procstats["sleep"]  +
						 procstats["wio"]    +
						 procstats["zombie"] +
						 procstats["stop"]   +
						 procstats["swap"]

	return procstats, nil
}
