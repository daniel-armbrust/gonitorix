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
	"time"
	
	"gonitorix/internal/procfs"
	"gonitorix/internal/logging"
)

// computeTotalCPUTimes calculates the total CPU times  by summing all 
// execution states provided by /proc/stat. 
func computeTotalCPUTimes(times *procfs.CPUTimes) uint64 {
	if times == nil {
		return 0
	}

	total := times.User + times.Nice + times.System + times.Idle +
		     times.IOWait + times.IRQ + times.SoftIRQ +	times.Steal +
		     times.Guest

	if logging.DebugEnabled() {
		logging.Debug("PROCESS", "Total CPU ticks: %d", total)
	}

	return total
}

// computeDeltaT returns the elapsed time in seconds since the last
// collection cycle. On first execution, it returns the configured step.
func computeDeltaT(step int) float64 {
	now := float64(time.Now().UnixNano()) / 1e9

	delta := float64(step)
	
	if lastTimestamp != 0 {
		delta = now - lastTimestamp
	}

	lastTimestamp = now
	
	return delta
}

func computeDiskBytes(procName string, diskBytes uint64, deltaT float64) float64 {
	prev, exists := processHistory[procName]

	if !exists {
		processHistory[procName] = &processStat{
			diskBytes: diskBytes,
		}
		return 0
	}

	var delta uint64

	if diskBytes >= prev.diskBytes {
		delta = diskBytes - prev.diskBytes
	}

	prev.diskBytes = diskBytes

	if deltaT <= 0 {
		return 0
	}

	return float64(delta) / deltaT
}

// computeNetBytes calculates the estimated network I/O rate (bytes per second)
// based on cumulative non-disk I/O counters.
func computeNetBytes(procName string, netBytes uint64, deltaT float64) float64 {
	prev, exists := processHistory[procName]

	if !exists {
		processHistory[procName] = &processStat{
			netBytes: netBytes,
		}
		return 0
	}

	var delta uint64

	if netBytes >= prev.netBytes {
		delta = netBytes - prev.netBytes
	}

	prev.netBytes = netBytes

	if deltaT <= 0 {
		return 0
	}

	return float64(delta) / deltaT
}

// computeVCS calculates the voluntary context switch rate
// (switches per second) for a given process.
func computeVCS(procName string, vcs uint64, deltaT float64) float64 {
	prev, exists := processHistory[procName]

	if !exists {
		processHistory[procName] = &processStat{
			vcs: vcs,
		}
		return 0
	}

	var delta uint64

	if vcs >= prev.vcs {
		delta = vcs - prev.vcs
	}

	prev.vcs = vcs

	if deltaT <= 0 {
		return 0
	}

	return float64(delta) / deltaT
}

// computeICS calculates the involuntary context switch rate
// (switches per second) for a given process.
func computeICS(procName string, ics uint64, deltaT float64) float64 {
	prev, exists := processHistory[procName]

	if !exists {
		processHistory[procName] = &processStat{
			ics: ics,
		}
		return 0
	}

	var delta uint64
	
	if ics >= prev.ics {
		delta = ics - prev.ics
	} else {
		// Counter reset or process restart
		prev.ics = ics
		return 0
	}

	prev.ics = ics

	if deltaT <= 0 {
		return 0
	}

	return float64(delta) / deltaT
}

// computeCPU calculates the CPU usage percentage of a process
// using cumulative process CPU ticks and total system CPU ticks.
func computeCPU(procName string, totalCPUTimes uint64, aggCPU uint64, _ float64) float64 {
	prev, exists := processHistory[procName]

	if !exists {
		processHistory[procName] = &processStat{
			totalCPU:  aggCPU,
			systemCPU: totalCPUTimes,
		}
		return 0
	}

	// Protect against counter reset
	if aggCPU < prev.totalCPU || totalCPUTimes < prev.systemCPU {
		prev.totalCPU = aggCPU
		prev.systemCPU = totalCPUTimes
		return 0
	}

	deltaProc := aggCPU - prev.totalCPU
	deltaSys  := totalCPUTimes - prev.systemCPU

	prev.totalCPU = aggCPU
	prev.systemCPU = totalCPUTimes

	if deltaSys == 0 {
		return 0
	}

	return 100.0 * float64(deltaProc) / float64(deltaSys)
}

// computeUptime calculates the process uptime in seconds
// and prevents regression if the computed value decreases.
func computeAggregatedUptime(currentMax float64, startTimeTicks uint64, sysUptime float64, ticksPerSecond uint64) float64 {
	if ticksPerSecond == 0 {
		return currentMax
	}

	startTimeSeconds := float64(startTimeTicks) / float64(ticksPerSecond)
	diff := sysUptime - startTimeSeconds

	if diff > currentMax {
		return diff
	}

	return currentMax
}