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
	"context"

	"gonitorix/internal/procfs"
	"gonitorix/internal/logging"
)

func measure(ctx context.Context) {
	memory, err := procfs.ReadMemory(ctx)

	if err != nil {
		logging.Error("SYSTEM", "Cannot read /proc/meminfo: %v", err)
	}

	loadAvg, err := procfs.ReadLoadAvg(ctx)

	if err != nil {
		logging.Error("SYSTEM", "Cannot read /proc/loadavg: %v", err)
	}

	entropy, err := procfs.ReadEntropy(ctx)

	if err != nil {
		logging.Error("SYSTEM", "Cannot read /proc/sys/kernel/random/entropy_avail: %v", err)
	}

	procInfo, err := procfs.ReadProcessStateCounts(ctx)

	if err != nil {
		logging.Error("SYSTEM", "Cannot read process state counts from /proc: %v", err)
	}

	uptime, err := procfs.ReadSystemUptime(ctx)

	if err != nil {
		logging.Error("SYSTEM", "Cannot read /proc/uptime: %v", err)
	}

	err = updateRRD(ctx, memory, loadAvg, entropy, procInfo, uptime)
	
	if err != nil {
		logging.Error("SYSTEM", "RRD update failed: %v", err)
	}
}