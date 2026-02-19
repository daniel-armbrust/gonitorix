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

import (
	"syscall"

	"gonitorix/internal/logging"
)

// getFilesystemUsage returns the percentage of used disk space for the
// given mount point using syscall.Statfs.
//
// The calculation is:
//
//     (total - available) / total * 100
//
// Bavail is used to match the behavior of tools like `df`.
// Returns 0 on error or if total size is zero.
func getFilesystemUsage(mountPoint string) float64 {
	var stat syscall.Statfs_t

	err := syscall.Statfs(mountPoint, &stat)

	if err != nil {
		logging.Warn("FILESYSTEM", "Stats failed for %s: %v", mountPoint, err,)
		return 0
	}

	total := float64(stat.Blocks) * float64(stat.Bsize)
	free := float64(stat.Bavail) * float64(stat.Bsize)

	if total == 0 {
		return 0
	}

	used := total - free
	usage := (used / total) * 100.0

	return usage
}

// getFilesystemInodeUsage returns the percentage of used inodes for the
// given mount point using syscall.Statfs.
//
// The calculation is:
//
//     (total - free) / total * 100
//
// Returns 0 on error or if the filesystem reports zero total inodes.
func getFilesystemInodeUsage(mountPoint string) float64 {
	var stat syscall.Statfs_t

	err := syscall.Statfs(mountPoint, &stat)

	if err != nil {
		logging.Warn("FILESYSTEM", "Stats (inode) failed for %s: %v", mountPoint, err,)
		return 0
	}

	total := float64(stat.Files)
	free := float64(stat.Ffree)

	if total == 0 {
		return 0
	}

	used := total - free
	usage := (used / total) * 100.0

	return usage
}