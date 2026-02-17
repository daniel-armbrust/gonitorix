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
	"os"
	"strconv"
	"bufio"
	"strings"
	"context"
	"fmt"
	
	"gonitorix/internal/logging"
)

// ReadProcStat reads /proc/stat and returns cumulative global CPU time counters
func ReadProcStat(ctx context.Context) (*ProcStat, error) {
	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading /proc/stat")
	}

	select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
	}

	file, err := os.Open("/proc/stat")

	if err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Failed to open /proc/stat: %v", err)
		}
		return nil, err
	}
	defer file.Close()

	ps := &ProcStat{}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// Check cancellation during scan
		if err := ctx.Err(); err != nil {
			if logging.DebugEnabled() {
				logging.Debug("PROCFS", "Context cancelled while reading /proc/stat")
			}
			return nil, err
		}

		line := scanner.Text()

		// -----------------------------------------
		// cpu line
		// -----------------------------------------
		if strings.HasPrefix(line, "cpu ") {
			if logging.DebugEnabled() {
				logging.Debug("PROCFS", "CPU line: %s", line)
			}

			fields := strings.Fields(line)

			if len(fields) >= 10 {
				ps.User, _ = strconv.ParseUint(fields[1], 10, 64)
				ps.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
				ps.System, _ = strconv.ParseUint(fields[3], 10, 64)
				ps.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
				ps.IOWait, _ = strconv.ParseUint(fields[5], 10, 64)
				ps.IRQ, _ = strconv.ParseUint(fields[6], 10, 64)
				ps.SoftIRQ, _ = strconv.ParseUint(fields[7], 10, 64)
				ps.Steal, _ = strconv.ParseUint(fields[8], 10, 64)
				ps.Guest, _ = strconv.ParseUint(fields[9], 10, 64)

				if logging.DebugEnabled() {
					logging.Debug(
						"PROCFS",
						"CPU parsed user=%d nice=%d sys=%d idle=%d iowait=%d irq=%d sirq=%d steal=%d guest=%d",
						ps.User,
						ps.Nice,
						ps.System,
						ps.Idle,
						ps.IOWait,
						ps.IRQ,
						ps.SoftIRQ,
						ps.Steal,
						ps.Guest,
					)
				}

			} else if logging.DebugEnabled() {
				logging.Debug("PROCFS", "CPU line has insufficient fields (%d)", len(fields))
			}

			continue
		}

		// -----------------------------------------
		// context switches
		// -----------------------------------------
		if strings.HasPrefix(line, "ctxt ") {
			if logging.DebugEnabled() {
				logging.Debug("PROCFS", "CTXT line: %s", line)
			}

			fields := strings.Fields(line)

			if len(fields) == 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid ctxt value: %w", err)
				}

				ps.ContextSwitches = val

				if logging.DebugEnabled() {
					logging.Debug("PROCFS", "Context switches: %d", ps.ContextSwitches)
				}
			}

			continue
		}

		// -----------------------------------------
		// processes (forks)
		// -----------------------------------------
		if strings.HasPrefix(line, "processes ") {
			if logging.DebugEnabled() {
				logging.Debug("PROCFS", "Processes line: %s", line)
			}

			fields := strings.Fields(line)

			if len(fields) == 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid processes value: %w", err)
				}

				ps.Forks = val
				ps.Vforks = 0 // Not exposed in modern kernels

				if logging.DebugEnabled() {
					logging.Debug("PROCFS", "Forks: %d", ps.Forks)
				}
			}

			continue
		}
	}

	if err := scanner.Err(); err != nil {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Scanner error: %v", err)
		}

		return nil, err
	}

	return ps, nil
}

// ReadProcDentryStat reads /proc/sys/fs/dentry-state and returns filesystem
// dentry cache statistics.
func ReadProcDentryStat(ctx context.Context) (*ProcDentryStat, error) {
	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading dentry/file/inode raw stats from /proc/sys/fs")
	}

	select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
	}

	stats := &ProcDentryStat{}

	// --------------------------------------------------
	// /proc/sys/fs/dentry-state
	// --------------------------------------------------
	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading /proc/sys/fs/dentry-state")
	}

	data, err := os.ReadFile("/proc/sys/fs/dentry-state")
	if err != nil {
		return nil, fmt.Errorf("cannot read dentry-state: %w", err)
	}

	fields := strings.Fields(string(data))

	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid dentry-state format")
	}

	stats.DentryUsed, err = strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid dentry used value: %w", err)
	}

	stats.DentryUnused, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid dentry unused value: %w", err)
	}

	if logging.DebugEnabled() {
		logging.Debug(
			"PROCFS",
			"Dentry raw used=%d unused=%d",
			stats.DentryUsed,
			stats.DentryUnused,
		)
	}

	// --------------------------------------------------
	// /proc/sys/fs/file-nr
	// --------------------------------------------------
	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading /proc/sys/fs/file-nr")
	}

	data, err = os.ReadFile("/proc/sys/fs/file-nr")
	if err != nil {
		return nil, fmt.Errorf("cannot read file-nr: %w", err)
	}

	fields = strings.Fields(string(data))

	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid file-nr format")
	}

	stats.FileUsed, err = strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid file used value: %w", err)
	}

	stats.FileMax, err = strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid file max value: %w", err)
	}

	if logging.DebugEnabled() {
		logging.Debug(
			"PROCFS",
			"File raw used=%d max=%d",
			stats.FileUsed,
			stats.FileMax,
		)
	}

	// --------------------------------------------------
	// /proc/sys/fs/inode-nr
	// --------------------------------------------------
	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading /proc/sys/fs/inode-nr")
	}

	data, err = os.ReadFile("/proc/sys/fs/inode-nr")
	if err != nil {
		return nil, fmt.Errorf("cannot read inode-nr: %w", err)
	}

	fields = strings.Fields(string(data))

	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid inode-nr format")
	}

	stats.InodeUsed, err = strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid inode used value: %w", err)
	}

	stats.InodeUnused, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid inode unused value: %w", err)
	}

	if logging.DebugEnabled() {
		logging.Debug(
			"PROCFS",
			"Inode raw used=%d unused=%d",
			stats.InodeUsed,
			stats.InodeUnused,
		)
	}

	return stats, nil
}

// ReadCPUTimes reads the aggregate CPU time counters from /proc/stat
// and returns the raw cumulative jiffy values for each CPU state.
func ReadCPUTimes(ctx context.Context) (*CPUTimes, error) {
	const path = "/proc/stat"

	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading %s", path)
	}

	// Fast cancel
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

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		line := scanner.Text()

		// /proc/stat
		//    - cpu = The numbers represent the amount of time the CPU has 
		//            spent performing different types of work.
		//           
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)

			if len(fields) < 10 {
				return nil, fmt.Errorf("invalid cpu line format")
			}

			// user: normal processes executing in user mode
			// nice: niced processes executing in user mode
			// system: processes executing in kernel mode
			// idle: cumulative time that the CPU has spent idle since system startup.
			// iowait: waiting for I/O to complete
			// irq: servicing interrupts
			// sirq: servicing softirqs
			// steal: involuntary wait
			user, _   := strconv.ParseUint(fields[1], 10, 64)
			nice, _   := strconv.ParseUint(fields[2], 10, 64)
			system, _ := strconv.ParseUint(fields[3], 10, 64)
			idle, _   := strconv.ParseUint(fields[4], 10, 64)
			iowait, _ := strconv.ParseUint(fields[5], 10, 64)
			irq, _    := strconv.ParseUint(fields[6], 10, 64)
			sirq, _   := strconv.ParseUint(fields[7], 10, 64)
			steal, _  := strconv.ParseUint(fields[8], 10, 64)

			var guest uint64
			if len(fields) > 9 {
				guest, _ = strconv.ParseUint(fields[9], 10, 64)
			}

			if logging.DebugEnabled() {
				logging.Debug(
					"PROCFS",
					"CPU raw times user=%d nice=%d sys=%d idle=%d iowait=%d irq=%d sirq=%d steal=%d guest=%d",
					user, nice, system, idle, iowait, irq, sirq, steal, guest,
				)
			}

			return &CPUTimes{
				User:    user,
				Nice:    nice,
				System:  system,
				Idle:    idle,
				IOWait:  iowait,
				IRQ:     irq,
				SoftIRQ: sirq,
				Steal:   steal,
				Guest:   guest,
			}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("cpu line not found in /proc/stat")
}
