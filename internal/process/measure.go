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

import (
	"context"
	"os"

	"gonitorix/internal/config"
	"gonitorix/internal/procfs"
	"gonitorix/internal/logging"
)

func measure(ctx context.Context) {
	// -------------------------------------------------
	// 1. Discover running PIDs
	// -------------------------------------------------
	procPids, err := findProcessPIDs(ctx)

	if err != nil {
		logging.Error("PROCESS", "No running PIDs found for configured processes to be monitored.")
		return
	}

	if len(procPids) == 0 {
		logging.Warn("PROCESS", "No running PIDs found for configured processes.")
		return
	}

	// -------------------------------------------------
	// 2. Compute deltaT (global for this cycle)
	// -------------------------------------------------

	deltaT := computeDeltaT(config.ProcessCfg.Step)

	// -------------------------------------------------
	// 3. Dependencies
	// -------------------------------------------------
	
	cpuTimes, err := procfs.ReadCPUTimes(ctx)

	if err != nil {
		logging.Error("PROCESS", "Cannot read system CPU times: %v", err)
		return
	}

	// $s_usage
	totalCPUTimes := computeTotalCPUTimes(cpuTimes)

	sysUptime, err := procfs.ReadSystemUptime(ctx)

	if err != nil {
		logging.Error("PROCESS", "Cannot read system uptime: %v", err)
		return
	}

	ticksPerSecond, err := procfs.GetClockTicks(ctx)

	if err != nil {
		logging.Error("PROCESS", "Cannot read clock ticks: %v", err)
		return
	}	

	// -------------------------------------------------
	// 4. Process each configured process
	// -------------------------------------------------

	for procName, pids := range procPids {
		if len(pids) == 0 {
			logging.Warn("PROCESS", "No running PIDs for process %q. Skipping...", procName)
			continue
		}

		logging.Debug("PROCESS", "Process %s has %d PIDs", procName, len(pids))

		// Total PIDs.
		proCount := float64(len(pids))

		// System page size (bytes), used to convert RSS pages to bytes.
		pageSize := uint64(os.Getpagesize())

		// ---------------------------------------------
		// Aggregate per PID
		// ---------------------------------------------

		var agg aggregatedProcessStat	

		for _, pid := range pids {
			// /proc/<pid>/stat
			procStat, err := procfs.ReadProcessStat(ctx, pid)

			if err != nil {
				continue
			}

			if procStat.RSS > 0 {
				rssBytes := uint64(procStat.RSS) * pageSize
				agg.memoryBytes += rssBytes
			}

			agg.threads += procStat.Threads
			agg.cpu += procStat.UTime + procStat.STime // $p_usage
			agg.uptime = computeAggregatedUptime(agg.uptime, procStat.StartTime, sysUptime, ticksPerSecond)

			// /proc/<pid>/io
			procIOStat, err := procfs.ReadProcessIOStat(ctx, pid)

			if err == nil {
				// Physical disk I/O
				agg.diskBytes += procIOStat.ReadBytes + procIOStat.WriteBytes

				logical  := procIOStat.RChar + procIOStat.WChar
				physical := procIOStat.ReadBytes + procIOStat.WriteBytes

				if logical >= physical {
					agg.netBytes += logical - physical
				}
			}

			// /proc/<pid>/fd + ctx switches
			procFDCtxStat, err := procfs.ReadProcessFDAndCtxStat(ctx, pid)

			if err == nil {
				agg.vcs += procFDCtxStat.VoluntaryCtxSwitches
				agg.ics += procFDCtxStat.InvoluntaryCtxSwitches
				agg.openFDs += procFDCtxStat.OpenFDs
			}
		} 

		// -------------------------------------------------
		// 5. Compute deltas using history
		// -------------------------------------------------
		cpu := computeCPU(procName, totalCPUTimes, agg.cpu, deltaT)
		diskBytes := computeDiskBytes(procName, agg.diskBytes, deltaT)
		netBytes := computeNetBytes(procName, agg.netBytes, deltaT)
		vcs := computeVCS(procName, agg.vcs, deltaT)
		ics := computeICS(procName, agg.ics, deltaT)
	
		// -------------------------------------------------
		// 6. Update RRD
		// -------------------------------------------------
		err := updateRRD(ctx, procName, cpu, agg.memoryBytes, diskBytes, netBytes, 
			             float64(agg.openFDs), proCount, float64(agg.threads), 
						 vcs, ics, agg.uptime, 0)

		if err != nil {
			logging.Error("PROCESS", "RRD update failed for %q: %v", procName, err)
		}			
	}
}