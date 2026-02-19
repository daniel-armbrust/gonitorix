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
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
)

// ReadDiskStats reads and parses /proc/diskstats, returning all block devices
// reported by the kernel.
func ReadDiskStats(ctx context.Context) ([]DiskStat, error) {
	file, err := os.Open("/proc/diskstats")

	if err != nil {
		return nil, err
	}
	defer file.Close()

	var stats []DiskStat
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
		}

		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}

		major, err1 := strconv.ParseUint(fields[0], 10, 32)
		minor, err2 := strconv.ParseUint(fields[1], 10, 32)
		timeIO, err3 := strconv.ParseUint(fields[12], 10, 64)
		weightedIO, err4 := strconv.ParseUint(fields[13], 10, 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			continue
		}

		stats = append(stats, DiskStat{
			Major:               uint32(major),
			Minor:               uint32(minor),
			Device:              fields[2],
			TimeDoingIO:         timeIO,
			WeightedTimeDoingIO: weightedIO,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
