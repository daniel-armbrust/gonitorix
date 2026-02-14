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
	"fmt"
	"os"
	"strings"
	"bufio"
	"strconv"

	"gonitorix/internal/utils"
	"gonitorix/internal/logging"
)

var clockTicks uint64

// GetClockTicks retrieves the system clock ticks per second (HZ value)
// via "getconf CLK_TCK". The result is cached to avoid repeated system calls.
func GetClockTicks(ctx context.Context) (uint64, error) {
	if clockTicks != 0 {
		return clockTicks, nil
	}

	select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
	}

	out, err := utils.ExecCommandOutput(ctx, "PROCFS", "getconf", "CLK_TCK")
	if err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Failed to execute getconf CLK_TCK: %v", err)
		}
		return 0, fmt.Errorf("cannot get CLK_TCK: %w", err)
	}

	out = strings.TrimSpace(out)

	val, err := strconv.ParseUint(out, 10, 64)
	if err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Invalid CLK_TCK value '%s': %v", out, err)
		}
		return 0, fmt.Errorf("invalid CLK_TCK value: %w", err)
	}

	if val == 0 {
		return 0, fmt.Errorf("invalid CLK_TCK value: %d", val)
	}

	clockTicks = val

	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "CLK_TCK detected: %d", clockTicks)
	}

	return clockTicks, nil
}

// ReadSystemUptime reads /proc/uptime and returns the system uptime in seconds.
func ReadSystemUptime(ctx context.Context) (float64, error) {
	file, err := os.Open("/proc/uptime")

	if err != nil {
		logging.Error("PROCFS", "Cannot read /proc/uptime: %v",	err,)
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

			logging.Warn("PROCFS", "Failed to parse uptime value: %s", line,)
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("PROCFS", "Error reading /proc/uptime: %v", err,)
		return 0, err
	}

	return 0, fmt.Errorf("uptime not found")
}