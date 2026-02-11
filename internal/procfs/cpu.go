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
				ps.User, _ = strconv.ParseFloat(fields[1], 64)
				ps.Nice, _ = strconv.ParseFloat(fields[2], 64)
				ps.Sys, _ = strconv.ParseFloat(fields[3], 64)
				ps.Idle, _ = strconv.ParseFloat(fields[4], 64)
				ps.Iowait, _ = strconv.ParseFloat(fields[5], 64)
				ps.IRQ, _ = strconv.ParseFloat(fields[6], 64)
				ps.SIRQ, _ = strconv.ParseFloat(fields[7], 64)
				ps.Steal, _ = strconv.ParseFloat(fields[8], 64)
				ps.Guest, _ = strconv.ParseFloat(fields[9], 64)

				if logging.DebugEnabled() {
					logging.Debug(
						"PROCFS",
						"CPU parsed user=%.0f nice=%.0f sys=%.0f idle=%.0f iowait=%.0f irq=%.0f sirq=%.0f steal=%.0f guest=%.0f",
						ps.User,
						ps.Nice,
						ps.Sys,
						ps.Idle,
						ps.Iowait,
						ps.IRQ,
						ps.SIRQ,
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
				ps.ContextSwitches, _ = strconv.ParseInt(fields[1], 10, 64)

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
				ps.Forks, _ = strconv.ParseInt(fields[1], 10, 64)
				ps.Vforks = 0

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
		logging.Debug("PROCFS", "Reading dentry/file/inode stats from /proc/sys/fs")
	}

	// Fast cancel check
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

	if data, err := os.ReadFile("/proc/sys/fs/dentry-state"); err == nil {
		fields := strings.Fields(string(data))

		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "dentry-state fields: %v", fields)
		}

		if len(fields) >= 2 {
			a, _ := strconv.ParseFloat(fields[0], 64)
			b, _ := strconv.ParseFloat(fields[1], 64)

			if a+b > 0 {
				stats.Dentry = (a * 100) / (a + b)

				if logging.DebugEnabled() {
					logging.Debug("PROCFS", "dentry usage: a=%.0f b=%.0f usage=%.2f%%", a, b, stats.Dentry,)
				}
			}

		} else if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Invalid dentry-state format (%d fields)", len(fields))
		}
	} else {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Failed to read /proc/sys/fs/dentry-state: %v", err)
		}

		return nil, err
	}

	// Context check between reads
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// --------------------------------------------------
	// /proc/sys/fs/file-nr
	// --------------------------------------------------
	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading /proc/sys/fs/file-nr")
	}

	if data, err := os.ReadFile("/proc/sys/fs/file-nr"); err == nil {
		fields := strings.Fields(string(data))

		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "file-nr fields: %v", fields)
		}

		if len(fields) >= 3 {
			used, _ := strconv.ParseFloat(fields[0], 64)
			max, _ := strconv.ParseFloat(fields[2], 64)

			if max > 0 {
				stats.File = (used * 100) / max

				if logging.DebugEnabled() {
					logging.Debug("PROCFS",	"file usage: used=%.0f max=%.0f usage=%.2f%%", used, max, stats.File,)
				}
			}

		} else if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Invalid file-nr format (%d fields)", len(fields))
		}

	} else {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Failed to read /proc/sys/fs/file-nr: %v", err)
		}

		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// --------------------------------------------------
	// /proc/sys/fs/inode-nr
	// --------------------------------------------------
	if logging.DebugEnabled() {
		logging.Debug("PROCFS", "Reading /proc/sys/fs/inode-nr")
	}

	if data, err := os.ReadFile("/proc/sys/fs/inode-nr"); err == nil {
		fields := strings.Fields(string(data))

		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "inode-nr fields: %v", fields)
		}

		if len(fields) >= 2 {
			a, _ := strconv.ParseFloat(fields[0], 64)
			b, _ := strconv.ParseFloat(fields[1], 64)

			if a+b > 0 {
				stats.Inode = (a * 100) / (a + b)

				if logging.DebugEnabled() {
					logging.Debug("PROCFS", "inode usage: a=%.0f b=%.0f usage=%.2f%%", a, b, stats.Inode,)
				}
			}

		} else if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Invalid inode-nr format (%d fields)", len(fields))
		}
	} else {
		if logging.DebugEnabled() {
			logging.Debug("PROCFS", "Failed to read /proc/sys/fs/inode-nr: %v", err)
		}

		return nil, err
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

		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)

			if len(fields) < 10 {
				return nil, fmt.Errorf("invalid cpu line format")
			}

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