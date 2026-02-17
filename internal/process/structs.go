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
 
package process

type processStat struct {
	rrdFile	  string
	totalCPU  uint64
	systemCPU uint64 
	netBytes  uint64
	diskBytes uint64
	ics       uint64
	vcs       uint64
}

type aggregatedProcessStat struct {
	cpu         uint64
    memoryBytes uint64
	threads     int64
	uptime      float64
	diskBytes   uint64
	netBytes    uint64
	ics         uint64
    vcs         uint64
	openFDs     uint64
}