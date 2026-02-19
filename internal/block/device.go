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
 
package block

import (
	"os"
	"fmt"
	"syscall"

	"gonitorix/internal/logging"

	"golang.org/x/sys/unix"
)

// GetDeviceMajorMinor resolves a device path and returns its Linux major
// and minor numbers.
func GetDeviceMajorMinor(device string) (uint32, uint32, error) {
	if logging.DebugEnabled() {
		logging.Debug("BLOCK", "Resolving major/minor for device: %s", device,)
	}

	fi, err := os.Stat(device)

	if err != nil {
		logging.Error("BLOCK", "Stat failed for device %s: %v",	device, err,)
		return 0, 0, err
	}

	stat, ok := fi.Sys().(*syscall.Stat_t)
	
	if !ok {
		err := fmt.Errorf("invalid stat type for device %s", device)
		logging.Error("BLOCK", "%v", err)
		return 0, 0, err
	}

	major := unix.Major(uint64(stat.Rdev))
	minor := unix.Minor(uint64(stat.Rdev))

	if logging.DebugEnabled() {
		logging.Debug("BLOCK", "Device %s - major=%d minor=%d", device,	major, minor,)
	}

	return major, minor, nil
}
