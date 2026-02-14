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

// nanKernelStats returns a kernel statistics structure with CPU percentage
// fields set to NaN when delta computation is not possible (e.g., counter
// reset or invalid sample). Raw monotonic counters (context switches, forks)
// are preserved, and filesystem usage percentages are computed from the
// raw procfs values.
func nanKernelStats(procStat *procfs.ProcStat, dentry *procfs.ProcDentryStat) procStatDentryStat {
	// Calculate dentry percentage from raw values
	var dentryPercent float64

	dTotal := dentry.DentryUsed + dentry.DentryUnused
	if dTotal > 0 {
		dentryPercent = 100.0 * float64(dentry.DentryUsed) / float64(dTotal)
	} else {
		dentryPercent = math.NaN()
	}

	// Calculate file percentage
	var filePercent float64
	if dentry.FileMax > 0 {
		filePercent = 100.0 * float64(dentry.FileUsed) / float64(dentry.FileMax)
	} else {
		filePercent = math.NaN()
	}

	// Calculate inode percentage
	var inodePercent float64
	iTotal := dentry.InodeUsed + dentry.InodeUnused
	if iTotal > 0 {
		inodePercent = 100.0 * float64(dentry.InodeUsed) / float64(iTotal)
	} else {
		inodePercent = math.NaN()
	}

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

		dentry: dentryPercent,
		file:   filePercent,
		inode:  inodePercent,
	}
}

// readKernelStatsAndStoreHistory reads raw kernel counters from procfs,
// validates monotonic CPU values against the previous snapshot,
// computes CPU percentage distribution and filesystem usage,
// preserves raw monotonic counters, updates the internal history,
// and returns the computed kernel statistics for the current cycle.
func readKernelStatsAndStoreHistory(ctx context.Context) (*procStatDentryStat, error) {
	select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
	}

	// -------------------------------------------------
	// 1. Read raw data from /proc
	// -------------------------------------------------
	procStat, errProcStat := procfs.ReadProcStat(ctx)
	dentryStat, errDentryStat := procfs.ReadProcDentryStat(ctx)

	if errProcStat != nil || errDentryStat != nil {
		logging.Warn(
			"KERNEL",
			"Kernel stats collection failed (proc=%v dentry=%v)",
			errProcStat,
			errDentryStat,
		)
		return nil, fmt.Errorf("kernel stats collection failed")
	}

	stats := procStatDentryStat{}

	// -------------------------------------------------
	// 2. Validate monotonic CPU counters
	// -------------------------------------------------
	if procStat.User >= lastProcStat.User &&
		procStat.Nice >= lastProcStat.Nice &&
		procStat.System >= lastProcStat.System &&
		procStat.Idle >= lastProcStat.Idle &&
		procStat.IOWait >= lastProcStat.IOWait &&
		procStat.IRQ >= lastProcStat.IRQ &&
		procStat.SoftIRQ >= lastProcStat.SoftIRQ &&
		procStat.Steal >= lastProcStat.Steal &&
		procStat.Guest >= lastProcStat.Guest {

		userDelta := procStat.User - lastProcStat.User
		niceDelta := procStat.Nice - lastProcStat.Nice
		sysDelta := procStat.System - lastProcStat.System
		idleDelta := procStat.Idle - lastProcStat.Idle
		iowDelta := procStat.IOWait - lastProcStat.IOWait
		irqDelta := procStat.IRQ - lastProcStat.IRQ
		sirqDelta := procStat.SoftIRQ - lastProcStat.SoftIRQ
		stealDelta := procStat.Steal - lastProcStat.Steal
		guestDelta := procStat.Guest - lastProcStat.Guest

		total := userDelta + niceDelta + sysDelta + idleDelta +
			iowDelta + irqDelta + sirqDelta + stealDelta +
			guestDelta

		if total > 0 {

			stats.user = 100.0 * float64(userDelta) / float64(total)
			stats.nice = 100.0 * float64(niceDelta) / float64(total)
			stats.sys = 100.0 * float64(sysDelta) / float64(total)
			stats.idle = 100.0 * float64(idleDelta) / float64(total)
			stats.iowait = 100.0 * float64(iowDelta) / float64(total)
			stats.irq = 100.0 * float64(irqDelta) / float64(total)
			stats.sirq = 100.0 * float64(sirqDelta) / float64(total)
			stats.steal = 100.0 * float64(stealDelta) / float64(total)
			stats.guest = 100.0 * float64(guestDelta) / float64(total)

		} else {
			stats = nanKernelStats(procStat, dentryStat)
		}

	} else {
		stats = nanKernelStats(procStat, dentryStat)
	}

	// -------------------------------------------------
	// 3. Raw monotonic counters
	// -------------------------------------------------
	stats.contextSwitches = procStat.ContextSwitches
	stats.forks = procStat.Forks
	stats.vforks = procStat.Vforks

	// -------------------------------------------------
	// 4. Compute filesystem usage percentages
	// -------------------------------------------------
	if total := dentryStat.DentryUsed + dentryStat.DentryUnused; total > 0 {
		stats.dentry = 100.0 * float64(dentryStat.DentryUsed) / float64(total)
	}

	if dentryStat.FileMax > 0 {
		stats.file = 100.0 * float64(dentryStat.FileUsed) / float64(dentryStat.FileMax)
	}

	if total := dentryStat.InodeUsed + dentryStat.InodeUnused; total > 0 {
		stats.inode = 100.0 * float64(dentryStat.InodeUsed) / float64(total)
	}

	// -------------------------------------------------
	// 5. Save history snapshot
	// -------------------------------------------------
	lastProcStat = *procStat

	return &stats, nil
}