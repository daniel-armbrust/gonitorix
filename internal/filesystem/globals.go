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
 
package filesystem

var (
	// lastTimestamp stores the timestamp of the previous polling cycle and is
	// used to compute the elapsed time (deltaT) between two reads.
	lastTimestamp float64

	// maxFilesystemsPerRRD defines the maximum number of filesystems that
	// can be stored within a single RRD file.
	maxFilesystemsPerRRD = 8

	// filesystemDevices stores runtime metadata for each monitored filesystem,
	// including device name, major/minor numbers and last I/O counters.
	filesystemDevices = map[string]*filesystemDevice{}
)