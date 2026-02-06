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