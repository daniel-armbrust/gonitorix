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

func initFilesystemMonitoring(ctx context.Context) {
	// Reset map in case of reload
	filesystemDevices = map[string]*filesystemDevice{}

	mounts, err := procfs.ReadMounts(ctx)

	if err != nil {
		logging.Error("FILESYSTEM",	"Unable to read mounts: %v", err,)
		return
	}

	// Index mounts by mountpoint for fast lookup
	mountMap := make(map[string]procfs.Mount)

	for _, m := range mounts {
		mountMap[m.MountPoint] = m
	}

	for i, mountPoint := range config.FilesystemCfg.MountPoints {
		select {
			case <-ctx.Done():
				return
			default:
		}

		m, ok := mountMap[mountPoint]

		if !ok {
			logging.Warn("FILESYSTEM", "Mount point '%s' not found", mountPoint,)
			continue
		}

		device := m.Device

		major, minor, err := block.GetDeviceMajorMinor(device)

		if err != nil {
			logging.Warn("FILESYSTEM", "Unable to resolve major/minor for '%s': %v", device, err,)
			continue
		}

		rrdIndex := i / maxFilesystemsPerRRD
		localIndex := i % maxFilesystemsPerRRD

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
			index:      localIndex,
			lastIOA:    0,
			lastTIM:    0,
		}

		if logging.DebugEnabled() {
			logging.Debug("FILESYSTEM",
				"Monitoring '%s' - %s (%d:%d) RRD=%s index=%d",
				mountPoint,
				device,
				major,
				minor,
				rrdFile,
				localIndex,
			)
		}
	}

	logging.Info("FILESYSTEM",
		"Filesystem monitoring initialized (%d mount points)",
		len(filesystemDevices),
	)
}
