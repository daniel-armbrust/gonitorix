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
 
package procfs

import (
	"os"
	"bufio"
	"fmt"
	"strings"
	"strconv"
	"context"

	"gonitorix/internal/logging"
)

// uptimeToString converts a system uptime value in seconds into a
// human-readable string representation.
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

// readUptime reads /proc/uptime and returns the system uptime in seconds.
func ReadUptime(ctx context.Context) (float64, error) {
	file, err := os.Open("/proc/uptime")

	if err != nil {
		logging.Error("SYSTEM", "Cannot read /proc/uptime: %v",	err,)
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	if scanner.Scan() {
		select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
		}

		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) >= 1 {
			val, err := strconv.ParseFloat(fields[0], 64)

			if err == nil {
				return val, nil
			}

			logging.Warn("SYSTEM", "Failed to parse uptime value: %s", line,)
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("SYSTEM", "Error reading /proc/uptime: %v", err,)
		return 0, err
	}

	return 0, fmt.Errorf("uptime not found")
}