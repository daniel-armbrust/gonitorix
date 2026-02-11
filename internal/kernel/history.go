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
	"context"

	"gonitorix/internal/procfs"
	"gonitorix/internal/logging"	
)

// nanKernelStats returns a ProcDentryStat with NaN values for all CPU
// percentage fields while preserving counter and VFS statistics.
// It is used when CPU deltas are invalid or cannot be computed.
func nanKernelStats(procStat *procfs.ProcStat, dentry *procfs.ProcDentryStat,) procStatDentryStat {
	return procStatDentryStat{
		user:   math.NaN(),
		nice:   math.NaN(),
		sys:    math.NaN(),
		idle:   math.NaN(),
		iowait: math.NaN(),
		irq:    math.NaN(),
		sirq:   math.NaN(),
		steal:  math.NaN(),
		guest:  math.NaN(),

		contextSwitches: procStat.ContextSwitches,
		forks:           procStat.Forks,
		vforks:          procStat.Vforks,

		dentry: dentry.Dentry,
		file:   dentry.File,
		inode:  dentry.Inode,
	}
}

// Reads kernel-related metrics from /proc, computes CPU usage percentages 
// and filesystem statistics, updates the historical snapshot used for delta 
// calculations, and writes the resulting values to the RRD database.
func readKernelStatsAndStoreHistory(ctx context.Context,) (*procStatDentryStat, error) {
	select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
	}

	// Read data from /proc
	procStat, errProcStat := procfs.ReadProcStat(ctx)
	dentryStat, errDentryStat := procfs.ReadProcDentryStat(ctx)

	if errProcStat != nil || errDentryStat != nil {
		logging.Warn("KERNEL", "Kernel stats collection failed (proc=%v dentry=%v)", errProcStat, errDentryStat,)
		return nil, fmt.Errorf("kernel stats collection failed")
	}

	stats := procStatDentryStat{}

	// Validate current CPU counters against the previous sample before
	// computing deltas.
	if procStat.User >= lastProcStat.User &&
		procStat.Nice >= lastProcStat.Nice &&
		procStat.Sys >= lastProcStat.Sys &&
		procStat.Idle >= lastProcStat.Idle &&
		procStat.Iowait >= lastProcStat.Iowait &&
		procStat.IRQ >= lastProcStat.IRQ &&
		procStat.SIRQ >= lastProcStat.SIRQ &&
		procStat.Steal >= lastProcStat.Steal &&
		procStat.Guest >= lastProcStat.Guest {

		userDelta := procStat.User - lastProcStat.User
		niceDelta := procStat.Nice - lastProcStat.Nice
		sysDelta := procStat.Sys - lastProcStat.Sys
		idleDelta := procStat.Idle - lastProcStat.Idle
		iowDelta := procStat.Iowait - lastProcStat.Iowait
		irqDelta := procStat.IRQ - lastProcStat.IRQ
		sirqDelta := procStat.SIRQ - lastProcStat.SIRQ
		stealDelta := procStat.Steal - lastProcStat.Steal
		guestDelta := procStat.Guest - lastProcStat.Guest

		total := userDelta + niceDelta + sysDelta + idleDelta +
			iowDelta + irqDelta + sirqDelta + stealDelta +
			guestDelta

		if total > 0 {
			stats = procStatDentryStat{
				user:   (userDelta * 100) / total,
				nice:   (niceDelta * 100) / total,
				sys:    (sysDelta * 100) / total,
				idle:   (idleDelta * 100) / total,
				iowait: (iowDelta * 100) / total,
				irq:    (irqDelta * 100) / total,
				sirq:   (sirqDelta * 100) / total,
				steal:  (stealDelta * 100) / total,
				guest:  (guestDelta * 100) / total,

				contextSwitches: procStat.ContextSwitches,
				forks:           procStat.Forks,
				vforks:          procStat.Vforks,

				dentry: dentryStat.Dentry,
				file:   dentryStat.File,
				inode:  dentryStat.Inode,
			}
		} else {
			stats = nanKernelStats(procStat, dentryStat)
		}
	} else {
		stats = nanKernelStats(procStat, dentryStat)
	}

	// Save history
	lastProcStat = *procStat

	return &stats, nil
}