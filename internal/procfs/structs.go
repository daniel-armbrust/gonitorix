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
 
package procfs

// -----------------------------------------------------
// /proc/net/dev (Network devices available and traffic)
// -----------------------------------------------------
type NetIfStat struct {	 
	RxBytes  float64
	TxBytes  float64	
	RxPkts   float64
	TxPkts   float64
	RxErrors float64
	TxErrors float64
}

// -----------------------------------------------------
// /proc/stat (CPU + global kernel counters)
// -----------------------------------------------------
type ProcStat struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	IOWait  uint64
	IRQ     uint64
	SoftIRQ uint64
	Steal   uint64
	Guest   uint64

	ContextSwitches uint64
	Forks           uint64
	Vforks          uint64
}

// -----------------------------------------------------
// /proc/sys/fs/* raw counters
// -----------------------------------------------------
type ProcDentryStat struct {
	DentryUsed   uint64
	DentryUnused uint64

	FileUsed uint64
	FileMax  uint64

	InodeUsed   uint64
	InodeUnused uint64
}

// -----------------------------------------------------
// /proc/<pid>/stat
// -----------------------------------------------------
type ProcessStat struct {
	PID        int
	UTime      uint64
	STime      uint64
	Threads    int64
	StartTime  uint64
	VSizeBytes uint64
}

// -----------------------------------------------------
// /proc/<pid>/io
// -----------------------------------------------------
type ProcessIOStat struct {
	RChar      uint64
	WChar      uint64
	ReadBytes  uint64
	WriteBytes uint64
}

// -----------------------------------------------------
// /proc/<pid>/fdinfo + /proc/<pid>/status
// -----------------------------------------------------
type ProcessFDAndCtxStat struct {
	OpenFDs                uint64
	VoluntaryCtxSwitches   uint64
	InvoluntaryCtxSwitches uint64
}

// -----------------------------------------------------
// /proc/stat → cpu line only
// -----------------------------------------------------
type CPUTimes struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	IOWait  uint64
	IRQ     uint64
	SoftIRQ uint64
	Steal   uint64
	Guest   uint64
}