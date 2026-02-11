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
	"strings"
	"fmt"
	"context"
	"regexp"
	"strconv"

	"gonitorix/internal/logging"
)

// ReadMemTotal reads /proc/meminfo and returns the total amount of
// system memory in kilobytes.
func ReadMemTotal(ctx context.Context) (uint64, error) {
	file, err := os.Open("/proc/meminfo")

	if err != nil {
		logging.Error("UTILS", "Cannot read /proc/meminfo: %v",	err,)
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)

			if len(fields) >= 2 {
				var val uint64

				if _, err := fmt.Sscanf(fields[1], "%d", &val); err != nil {
					return 0, err
				}

				return val, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("UTILS", "Error reading /proc/meminfo: %v", err,)
		return 0, err
	}

	return 0, fmt.Errorf("MemTotal not found")
}

// ReadMemory reads /proc/meminfo and returns selected memory statistics
// such as total, free, buffers, cache and active/inactive pages.
// The operation can be cancelled through the provided context.
func ReadMemory(ctx context.Context) (map[string]uint64, error) {
	file, err := os.Open("/proc/meminfo")

	if err != nil {
		logging.Error("SYSTEM", "Cannot read /proc/meminfo: %v", err,)
		return nil, err
	}
	defer file.Close()

	re := regexp.MustCompile(`^(\w+):\s+(\d+)\s+kB`)

	mem := make(map[string]uint64)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
		}

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
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("SYSTEM", "Error reading /proc/meminfo: %v", err,)
		return nil, err
	}

	// Include reclaimable and unreclaimable slabs into MemFree
	// so they are accounted for when calculating used memory.
	if srecl, ok := mem["SReclaimable"]; ok {
		mem["MemFree"] += srecl
	}

	if sun, ok := mem["SUnreclaim"]; ok {
		mem["MemFree"] += sun
	}

	if len(mem) == 0 {
		return nil, fmt.Errorf("no memory statistics collected")
	}

	if logging.DebugEnabled() {
		logging.Debug("SYSTEM",	"Collected memory statistics: %+v", mem,)
	}

	return mem, nil
}