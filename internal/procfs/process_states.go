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

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"strconv"

	"gonitorix/internal/logging"
)

// ReadProcessStateCounts scans /proc and counts processes by execution state,
// returning totals for running, sleeping, waiting for I/O, zombie,
// stopped and swapped processes.
func ReadProcessStateCounts(ctx context.Context) (map[string]uint64, error) {
	procstats := map[string]uint64{
		"run":    0,
		"sleep":  0,
		"wio":    0,
		"zombie": 0,
		"stop":   0,
		"swap":   0,
	}

	dirs, err := filepath.Glob("/proc/[0-9]*")

	if err != nil {
		logging.Error("PROCFS", "Failed to list /proc entries: %v",	err,)
		return nil, err
	}

	for _, dir := range dirs {
		select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
		}

		info, err := os.Stat(dir)

		if err != nil || !info.IsDir() {
			continue
		}

		statusFile := dir + "/status"

		if _, err := os.Stat(statusFile); err != nil {
			continue
		}

		f, err := os.Open(statusFile)

		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "State:") {
				fields := strings.Fields(line)

				if len(fields) >= 2 {
					state := fields[1]

					switch state {
						case "R":
							procstats["run"]++
						case "S":
							procstats["sleep"]++
						case "D":
							procstats["wio"]++
						case "Z":
							procstats["zombie"]++
						case "T":
							procstats["stop"]++
						case "W":
							procstats["swap"]++
					}
				}

				break
			}
		}

		f.Close()
	}

	procstats["total"] = procstats["run"] + procstats["sleep"] + procstats["wio"] +
			             procstats["zombie"] + procstats["stop"] + procstats["swap"]

	if len(procstats) == 0 {
		return nil, fmt.Errorf("no process statistics collected")
	}

	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Collected process states: %+v", procstats,)
	}

	return procstats, nil
}

// ReadProcessStat parses /proc/<pid>/stat and returns raw kernel counters
// for the given process.
func ReadProcessStat(ctx context.Context, pid int) (*ProcessStat, error) {
	path := fmt.Sprintf("/proc/%d/stat", pid)

	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading %s", path)
	}

	select {
		case <-ctx.Done():
			if logging.DebugEnabled() {
				logging.Debug("PROCFS", "Context cancelled before reading %s", path)
			}
			return nil, ctx.Err()
		default:
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Failed to read %s: %v", path, err)
		}
		return nil, err
	}

	line := strings.TrimSpace(string(data))

	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Raw stat line (pid=%d): %s", pid, line)
	}

	// pid (comm with spaces) state rest...
	re := regexp.MustCompile(`^\d+\s+\(.*?\)\s+\S+\s+(.*)$`)
	m := re.FindStringSubmatch(line)

	if len(m) != 2 {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Invalid stat format for pid %d", pid)
		}
		return nil, fmt.Errorf("invalid /proc/%d/stat format", pid)
	}

	fields := strings.Fields(m[1])

	if len(fields) < 20 {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Short stat line for pid %d (fields=%d)", pid, len(fields))
		}
		return nil, fmt.Errorf("short stat line for pid %d", pid)
	}

	// utime: amount of CPU in user mode
	// stime: amount of CPU in kernel mode
	// numThreads: total number of threads in the process (including the main thread)
	// startTime: is the time the process started after system boot
	// vsize: total virtual memory size of the process in bytes
	utime, _ := strconv.ParseUint(fields[10], 10, 64)
	stime, _ := strconv.ParseUint(fields[11], 10, 64)
	numThreads, _ := strconv.ParseInt(fields[16], 10, 64)
	startTime, _ := strconv.ParseUint(fields[18], 10, 64)
	vsize, _ := strconv.ParseUint(fields[19], 10, 64)

	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "PID %d stat parsed utime=%d stime=%d threads=%d starttime=%d vsize=%d",
			pid, utime, stime, numThreads, startTime, vsize,
		)
	}

	return &ProcessStat{
		PID:        pid,
		UTime:      utime,
		STime:      stime,
		Threads:    numThreads,
		StartTime:  startTime,
		VSizeBytes: vsize,
	}, nil
}

// ReadProcessIO reads raw I/O counters from /proc/<pid>/io for the given PID.
// It parses rchar, wchar, read_bytes and write_bytes and returns their
// cumulative values since the process started.
func ReadProcessIOStat(ctx context.Context, pid int) (*ProcessIOStat, error) {
	path := fmt.Sprintf("/proc/%d/io", pid)

	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading %s", path)
	}

	select {
		case <-ctx.Done():
			if logging.DebugEnabled() {
				logging.Debug("PROCFS", "Context cancelled before reading %s", path)
			}
			return nil, ctx.Err()
		default:
	}

	file, err := os.Open(path)

	if err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Failed to open %s: %v", path, err)
		}
		return nil, err
	}
	defer file.Close()

	stat := &ProcessIOStat{}		

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			if logging.DebugEnabled() {
				logging.Debug("PROCFS", "Context cancelled while reading %s", path)
			}
			return nil, err
		}

		line := scanner.Text()

		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "PID %d IO line: %s", pid, line)
		}

		switch {
			case strings.HasPrefix(line, "rchar:"):
				fmt.Sscanf(line, "rchar: %d", &stat.RChar)

			case strings.HasPrefix(line, "wchar:"):
				fmt.Sscanf(line, "wchar: %d", &stat.WChar)

			case strings.HasPrefix(line, "read_bytes:"):
				fmt.Sscanf(line, "read_bytes: %d", &stat.ReadBytes)

			case strings.HasPrefix(line, "write_bytes:"):
				fmt.Sscanf(line, "write_bytes: %d", &stat.WriteBytes)
		}
	}

	if err := scanner.Err(); err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Scanner error for PID %d: %v", pid, err)
		}
		return nil, err
	}

	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "PID %d IO parsed rchar=%d wchar=%d read_bytes=%d write_bytes=%d",
			pid, stat.RChar, stat.WChar, stat.ReadBytes, stat.WriteBytes,
		)
	}

	return stat, nil
}

// ReadProcessFDAndCtxStat reads raw file descriptor and context switch counters
// for the given PID from /proc/<pid>/fdinfo and /proc/<pid>/status.
func ReadProcessFDAndCtxStat(ctx context.Context, pid int) (*ProcessFDAndCtxStat, error) {
	select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
	}

	stat := &ProcessFDAndCtxStat{}

	// --------------------------------------------------
	// Count open file descriptors
	// --------------------------------------------------
	fdPath := fmt.Sprintf("/proc/%d/fdinfo", pid)

	entries, err := os.ReadDir(fdPath)
	if err == nil {
		for _, entry := range entries {

			if err := ctx.Err(); err != nil {
				return nil, err
			}

			name := entry.Name()
			if len(name) > 0 && name[0] != '.' {
				stat.OpenFDs++
			}
		}

		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "PID %d open FDs: %d", pid, stat.OpenFDs)
		}

	} else {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Cannot read fdinfo for PID %d: %v", pid, err)
		}
		return nil, err
	}

	// --------------------------------------------------
	// Context switches
	// --------------------------------------------------
	statusPath := fmt.Sprintf("/proc/%d/status", pid)

	file, err := os.Open(statusPath)
	if err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Cannot open status for PID %d: %v", pid, err)
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "voluntary_ctxt_switches:") {
			fmt.Sscanf(line, "voluntary_ctxt_switches: %d", &stat.VoluntaryCtxSwitches)
		}

		if strings.HasPrefix(line, "nonvoluntary_ctxt_switches:") {
			fmt.Sscanf(line, "nonvoluntary_ctxt_switches: %d", &stat.InvoluntaryCtxSwitches)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if logging.DebugEnabled() {
		logging.Debug("PROCFS",	"PID %d ctx switches: voluntary=%d involuntary=%d",
			pid, stat.VoluntaryCtxSwitches, stat.InvoluntaryCtxSwitches,
		)
	}

	return stat, nil
}