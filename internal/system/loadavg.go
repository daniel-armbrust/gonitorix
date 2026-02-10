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
	"os"
	"regexp"
	"strconv"
	"fmt"

	"gonitorix/internal/logging"
)

// readLoadAvg reads /proc/loadavg and returns the system load averages
// for 1, 5 and 15 minutes.
func readLoadAvg(ctx context.Context) (map[string]float64, error) {
	file, err := os.Open("/proc/loadavg")

	if err != nil {
		logging.Error("SYSTEM", "Cannot read /proc/loadavg: %v", err,)
		return nil, err
	}
	defer file.Close()

	re := regexp.MustCompile(`^(\d+\.\d+)\s+(\d+\.\d+)\s+(\d+\.\d+)`)

	load := make(map[string]float64)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
		}

		line := scanner.Text()

		m := re.FindStringSubmatch(line)

		if len(m) == 4 {
			load1, err1 := strconv.ParseFloat(m[1], 64)
			load5, err5 := strconv.ParseFloat(m[2], 64)
			load15, err15 := strconv.ParseFloat(m[3], 64)

			if err1 != nil || err5 != nil || err15 != nil {
				logging.Warn("SYSTEM", "Failed to parse loadavg values: %s", line,)
				continue
			}

			load["load1"] = load1
			load["load5"] = load5
			load["load15"] = load15

			return load, nil
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("SYSTEM", "Error reading /proc/loadavg: %v", err,)
		return nil, err
	}

	return nil, fmt.Errorf("no load average values found")
}