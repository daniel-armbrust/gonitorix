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
	"context"
	"fmt"
	"time"

	"gonitorix/internal/procfs"
	"gonitorix/internal/logging"
)

func measure(ctx context.Context) {
	now := float64(time.Now().Unix())

	if lastTimestamp == 0 {
		lastTimestamp = now
		return
	}

	deltaT := now - lastTimestamp
	lastTimestamp = now

	if deltaT <= 0 {
		return
	}

	stats, err := procfs.ReadDiskStats(ctx)

	if err != nil {
		logging.Error("FILESYSTEM", "Unable to read /proc/diskstats: %v", err,)
		return
	}

	diskMap := make(map[string]procfs.DiskStat)

	for _, s := range stats {
		key := fmt.Sprintf("%d:%d", s.Major, s.Minor)
		diskMap[key] = s
	}

	groupedValues := map[string][]string{}

	for _, dev := range filesystemDevices {
		select {
			case <-ctx.Done():
				return
			default:
		}

		key := fmt.Sprintf("%d:%d", dev.major, dev.minor)
		stat, ok := diskMap[key]

		if !ok {
			continue
		}

		currentIOA := stat.TimeDoingIO
		currentTIM := stat.WeightedTimeDoingIO

		if dev.lastIOA == 0 {
			dev.lastIOA = currentIOA
			dev.lastTIM = currentTIM
			continue
		}

		deltaIOA := currentIOA - dev.lastIOA
		deltaTIM := currentTIM - dev.lastTIM

		dev.lastIOA = currentIOA
		dev.lastTIM = currentTIM

		ioaPerSec := float64(deltaIOA) / deltaT
		timPerSec := float64(deltaTIM) / deltaT

		usage := getFilesystemUsage(dev.mountPoint)
		inode := getFilesystemInodeUsage(dev.mountPoint)

		rrdata := fmt.Sprintf(
			"%.2f:%.2f:%.2f:%.2f",
			usage,
			ioaPerSec,
			timPerSec,
			inode,
		)

		groupedValues[dev.rrdFile] = append(groupedValues[dev.rrdFile], rrdata)
	}

	for rrdFile, rrdata := range groupedValues {
		select {
			case <-ctx.Done():
				return
			default:
		}

		if err := updateRRD(ctx, rrdFile, rrdata); err != nil {
			logging.Error("FILESYSTEM",	"RRD update failed for '%s': %v", rrdFile, err,)
		}
	}
}