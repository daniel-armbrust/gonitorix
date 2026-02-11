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

type ProcStat struct {
	User   float64
	Nice   float64
	Sys    float64
	Idle   float64
	Iowait float64
	IRQ    float64
	SIRQ   float64
	Steal  float64
	Guest  float64

	ContextSwitches int64
	Forks           int64
	Vforks          int64
}

type ProcDentryStat struct {
	Dentry float64
	File   float64
	Inode  float64
}

type NetIfStats struct {	 
	RxBytes  float64
	TxBytes  float64	
	RxPkts   float64
	TxPkts   float64
	RxErrors float64
	TxErrors float64
}

type ProcessStat struct {
	PID        int
	UTime      uint64
	STime      uint64
	Threads    int64
	StartTime  uint64
	VSizeBytes uint64
}

type ProcessIOStat struct {
	PID        int
	RChar      uint64
	WChar      uint64
	ReadBytes  uint64
	WriteBytes uint64

	DiskBytes uint64
	NetBytes  uint64
}

type ProcessFDStat struct {
	PID                    int
	OpenFDs                int
	VoluntaryCtxSwitches   uint64
	InvoluntaryCtxSwitches uint64
}

type CPUTimes struct {
	User   uint64
	Nice   uint64
	System uint64
	Idle   uint64
	IOWait uint64
	IRQ    uint64
	SoftIRQ uint64
	Steal  uint64
	Guest  uint64
}