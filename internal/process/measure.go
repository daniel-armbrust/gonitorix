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
	"bufio"
	"context"
	"path"
	"strconv"
	"strings"
	
	"gonitorix/internal/config"
	"gonitorix/internal/utils"
	"gonitorix/internal/procfs"
	"gonitorix/internal/logging"
)

// findProcessPIDs scans the process table and returns all PIDs whose
// command name or full command line matches the given pattern.
func findProcessPIDs(ctx context.Context) (map[string][]int, error) {
	out, err := utils.ExecCommandOutput(ctx, "PROCESS", "ps", "-eo", "pid,comm,args")

	if err != nil {
		return nil, err
	}

	results := make(map[string][]int)

	// Build lookup map from config
	cfgNames := make(map[string]struct{})

	for _, p := range config.ProcessCfg.Processes {
		name := strings.TrimSpace(p.Name)

		if name != "" {
			cfgNames[name] = struct{}{}
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(out))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "PID ") || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pidStr := fields[0]
		comm := fields[1]

		args := ""
		if len(fields) > 2 {
			args = fields[2]
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		exe := args
		if exe != "" {
			exe = path.Base(exe)
		}

		if _, ok := cfgNames[comm]; ok {
			results[comm] = append(results[comm], pid)
			continue
		}

		if exe != "" {
			if _, ok := cfgNames[exe]; ok {
				results[exe] = append(results[exe], pid)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// computeTotalCPUTicks calculates the total accumulated CPU ticks by
// summing all CPU time states (user, nice, system, idle, iowait, irq,
// softirq, steal and guest).
func computeTotalCPUTicks(times *procfs.CPUTimes) uint64 {
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

// computeProcessStat aggregates raw per-PID counters into calculated
// CPU, memory, thread, and uptime metrics.
func computeProcessStat(ctx context.Context, pid int, sysUptime float64, ticksPerSecond uint64,) (*aggregatedProcessStat, error) {
	if logging.DebugEnabled() {
		logging.Debug("PROCESS", "Computing stats for PID %d", pid)
	}

	select {
		case <-ctx.Done():
			if logging.DebugEnabled() {
				logging.Debug("PROCESS", "Context cancelled before computing PID %d", pid)
			}
			return nil, ctx.Err()
		default:
	}

	ps, err := procfs.ReadProcessStat(ctx, pid)

	if err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCESS", "Failed to read stat for PID %d: %v", pid, err)
		}
		return nil, err
	}

	if logging.DebugEnabled() {
		logging.Debug(
			"PROCESS",
			"PID %d raw stat utime=%d stime=%d threads=%d starttime=%d vsize=%d",
			pid,
			ps.UTime,
			ps.STime,
			ps.Threads,
			ps.StartTime,
			ps.VSizeBytes,
		)
	}

	// CPU usage (raw ticks)
	cpuUsage := ps.UTime + ps.STime

	// Memory (already in bytes)
	memBytes := ps.VSizeBytes

	// Threads (subtract main thread like Perl)
	threads := ps.Threads - 1
	if threads < 0 {
		threads = 0
	}

	// Convert start time from ticks to seconds
	startSeconds := float64(ps.StartTime) / float64(ticksPerSecond)

	// Process uptime
	uptime := sysUptime - startSeconds
	if uptime < 0 {
		uptime = 0
	}

	if logging.DebugEnabled() {
		logging.Debug(
			"PROCESS",
			"PID %d computed cpuTicks=%d memBytes=%d threads=%d uptime=%.2f",
			pid,
			cpuUsage,
			memBytes,
			threads,
			uptime,
		)
	}

	return &aggregatedProcessStat{
		cpuUsageTicks: cpuUsage,
		memoryBytes:   memBytes,
		threads:       threads,
		uptimeSeconds: uptime,
	}, nil
}