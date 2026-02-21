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
	"context"
	"os"
	"bufio"
	"strings"
	"fmt"
	"strconv"

	"gonitorix/internal/logging"
)

func ReadInterruptStat(ctx context.Context) (*InterruptStat, error) {
	file, err := os.Open("/proc/stat")

	if err != nil {
		logging.Error("PROCFS", "Cannot read /proc/stat: %v", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "intr ") {
			fields := strings.Fields(line)

			if len(fields) < 2 {
				return nil, fmt.Errorf("invalid intr line format")
			}

			total, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				logging.Warn("PROCFS", "Failed to parse intr total: %s", line)
				return nil, err
			}

			var irqs []uint64

			// Starting from the third field, the values represent individual 
			// IRQ counters.
			for _, f := range fields[2:] {
				v, err := strconv.ParseUint(f, 10, 64)

				if err != nil {
					logging.Warn("PROCFS", "Failed to parse IRQ value: %s", f)
					continue
				}
				irqs = append(irqs, v)
			}

			return &InterruptStat{
				Total: total,
				IRQs:  irqs,
			}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("PROCFS", "Error reading /proc/stat: %v", err)
		return nil, err
	}

	return nil, fmt.Errorf("intr line not found in /proc/stat")
}