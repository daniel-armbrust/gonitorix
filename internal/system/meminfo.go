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