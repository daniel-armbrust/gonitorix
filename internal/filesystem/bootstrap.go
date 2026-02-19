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
	"fmt"
	"context"
	"path/filepath"
	
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/procfs"
	"gonitorix/internal/block"
)

func initFilesystemMonitoring(ctx context.Context) error {
	if logging.DebugEnabled() {
		logging.Debug("FILESYSTEM", "Initializing filesystem monitoring")
	}

	mounts, err := procfs.ReadMounts(ctx)

	if err != nil {
		logging.Error("FILESYSTEM",	"Unable to read mounts: %v", err,)
		return err
	}

	filesystemDevices = map[string]*filesystemDevice{}

	for i, mountPoint := range config.FilesystemCfg.MountPoints {
		select {
			case <-ctx.Done():
				logging.Warn("FILESYSTEM", "Initialization cancelled by context")
				return ctx.Err()
			default:
		}

		var device string

		for _, m := range mounts {
			if m.MountPoint == mountPoint {
				device = m.Device
				break
			}
		}

		if device == "" {
			logging.Warn("FILESYSTEM", "Mountpoint '%s' not found in /proc/self/mounts", mountPoint,)
			continue
		}

		// resolve symlink (ex: /dev/disk/by-uuid)
		resolved, err := filepath.EvalSymlinks(device)
		if err == nil {
			device = resolved
		}

		major, minor, err := block.GetDeviceMajorMinor(device)

		if err != nil {
			logging.Warn("FILESYSTEM", "Cannot obtain major/minor for device '%s': %v", device, err,)
			continue
		}

		rrdIndex := i / maxFilesystemsPerRRD

		rrdFile := filepath.Join(
			config.GlobalCfg.RRDPath,
			fmt.Sprintf("%sfs-%d.rrd",
				config.GlobalCfg.RRDHostnamePrefix,
				rrdIndex,
			),
		)

		filesystemDevices[mountPoint] = &filesystemDevice{
			rrdFile:    rrdFile,
			mountPoint: mountPoint,
			device:     device,
			major:      major,
			minor:      minor,
			lastIOA:    0,
			lastTIM:    0,
		}

		if logging.DebugEnabled() {
			logging.Debug("FILESYSTEM",
				"Monitoring '%s' - %s (%d:%d) RRD=%s",
				mountPoint,
				device,
				major,
				minor,
				rrdFile,
			)
		}
	}

	if logging.DebugEnabled() {
		logging.Debug("FILESYSTEM",	"Filesystem monitoring initialized (%d mountpoints)", len(filesystemDevices),)
	}

	return nil
}
