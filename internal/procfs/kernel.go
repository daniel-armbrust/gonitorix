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
	"fmt"

	"gonitorix/internal/logging"
)

// ReadEntropy reads the current available kernel entropy value from
// /proc/sys/kernel/random/entropy_avail.
func ReadEntropy(ctx context.Context) (uint64, error) {
	file, err := os.Open("/proc/sys/kernel/random/entropy_avail")

	if err != nil {
		logging.Error("SYSTEM", "Cannot read entropy file: %v", err,)
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

		val, err := strconv.ParseUint(line, 10, 64)

		if err != nil {
			continue
		}

		return val, nil
	}

	if err := scanner.Err(); err != nil {
		logging.Error("SYSTEM", "Error reading entropy file: %v", err,)
		return 0, err
	}

	return 0, fmt.Errorf("no entropy value found")
}
