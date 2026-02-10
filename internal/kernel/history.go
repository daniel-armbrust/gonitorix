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

package kernel

import (
	"math"
	"fmt"

	"gonitorix/internal/logging"
)

// nanKernelStats returns a procDentryStateStat with NaN values for all CPU
// percentage fields while preserving counter and VFS statistics.
// It is used when CPU deltas are invalid or cannot be computed.
func nanKernelStats(procStat *procStat,	dentry *dentryState,) procDentryStateStat {
	return procDentryStateStat{
		user:   math.NaN(),
		nice:   math.NaN(),
		sys:    math.NaN(),
		idle:   math.NaN(),
		iowait: math.NaN(),
		irq:    math.NaN(),
		sirq:   math.NaN(),
		steal:  math.NaN(),
		guest:  math.NaN(),

		contextSwitches: procStat.contextSwitches,
		forks:           procStat.forks,
		vforks:          procStat.vforks,

		dentry: dentry.dentry,
		file:   dentry.file,
		inode:  dentry.inode,
	}
}

// Reads kernel-related metrics from /proc, computes CPU usage percentages 
// and filesystem statistics, updates the historical snapshot used for delta 
// calculations, and writes the resulting values to the RRD database.
func updateKernelStats() (*procDentryStateStat, error) {
	procStat, errProcStat := readProcStat()
	dentryStateStat, errDentryStateStat := readDentryStateStat()

	if errProcStat != nil || errDentryStateStat != nil {
		logging.Warn(
			"KERNEL",
			"Kernel stats collection failed (proc=%v dentry=%v)",
			errProcStat,
			errDentryStateStat,
		)
		return nil, fmt.Errorf("kernel stats collection failed")
	}

	stats := procDentryStateStat{}

	// Validate current CPU counters against the previous sample before
	// computing deltas.
	if procStat.user >= lastProcStat.user &&
		procStat.nice >= lastProcStat.nice &&
		procStat.sys >= lastProcStat.sys &&
		procStat.idle >= lastProcStat.idle &&
		procStat.iowait >= lastProcStat.iowait &&
		procStat.irq >= lastProcStat.irq &&
		procStat.sirq >= lastProcStat.sirq &&
		procStat.steal >= lastProcStat.steal &&
		procStat.guest >= lastProcStat.guest {

		userDelta := procStat.user - lastProcStat.user
		niceDelta := procStat.nice - lastProcStat.nice
		sysDelta := procStat.sys - lastProcStat.sys
		idleDelta := procStat.idle - lastProcStat.idle
		iowDelta := procStat.iowait - lastProcStat.iowait
		irqDelta := procStat.irq - lastProcStat.irq
		sirqDelta := procStat.sirq - lastProcStat.sirq
		stealDelta := procStat.steal - lastProcStat.steal
		guestDelta := procStat.guest - lastProcStat.guest

		total := userDelta + niceDelta + sysDelta + idleDelta +
			iowDelta + irqDelta + sirqDelta + stealDelta +
			guestDelta

		if total > 0 {
			stats = procDentryStateStat{
				user:   (userDelta * 100) / total,
				nice:   (niceDelta * 100) / total,
				sys:    (sysDelta * 100) / total,
				idle:   (idleDelta * 100) / total,
				iowait: (iowDelta * 100) / total,
				irq:    (irqDelta * 100) / total,
				sirq:   (sirqDelta * 100) / total,
				steal:  (stealDelta * 100) / total,
				guest:  (guestDelta * 100) / total,

				contextSwitches: procStat.contextSwitches,
				forks:           procStat.forks,
				vforks:          procStat.vforks,

				dentry: dentryStateStat.dentry,
				file:   dentryStateStat.file,
				inode:  dentryStateStat.inode,
			}
		} else {
			stats = nanKernelStats(procStat, dentryStateStat)
		}

	} else {
		stats = nanKernelStats(procStat, dentryStateStat)
	}

	// Save history
	lastProcStat = *procStat

	return &stats, nil
}