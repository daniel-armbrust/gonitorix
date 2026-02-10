/*
 * Gonitorix - a system and network monitoring tool
 * Copyright (C) 2026 Daniel Armbrust <darmbrust@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
 
package system

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gonitorix/internal/logging"
)

// readProcInfo scans /proc and counts processes by execution state,
// returning totals for running, sleeping, waiting for I/O, zombie,
// stopped and swapped processes.
func readProcInfo(ctx context.Context) (map[string]uint64, error) {
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
		logging.Error("SYSTEM", "Failed to list /proc entries: %v",	err,)
		return nil, err
	}

	for _, dir := range dirs {
		select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
		}

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

	procstats["total"] = procstats["run"] + procstats["sleep"] + procstats["wio"] +
			             procstats["zombie"] + procstats["stop"] + procstats["swap"]

	if len(procstats) == 0 {
		return nil, fmt.Errorf("no process statistics collected")
	}

	if logging.DebugEnabled() {
		logging.Debug("SYSTEM", "Collected process states: %+v", procstats,)
	}

	return procstats, nil
}
