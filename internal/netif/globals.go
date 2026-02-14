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
 
package netif

import "gonitorix/internal/procfs"

var (
	// lastTimestamp stores the timestamp of the previous polling cycle and is
	// used to compute the elapsed time (deltaT) between two reads.
	lastTimestamp float64

	// lastIfStats stores the previous statistics snapshot for each network
	// interface, used to compute deltas between successive polling cycles.
	lastNetIfStats = make(map[string]procfs.NetIfStat)
)