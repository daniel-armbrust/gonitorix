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
	"os"
	"strconv"
	"bufio"
	"strings"
)

// readProcStat reads /proc/stat and returns cumulative global CPU time counters
// used by Gonitorix to calculate CPU usage over time.
func readProcStat() (*procStat, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ps := &procStat{}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// -----------------------------------------
		// cpu line
		//   Amount of time the CPU has spent performing different kinds of work.
		// -----------------------------------------
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)

			if len(fields) >= 10 {
				ps.user, _   = strconv.ParseFloat(fields[1], 64)
				ps.nice, _   = strconv.ParseFloat(fields[2], 64)
				ps.sys, _    = strconv.ParseFloat(fields[3], 64)
				ps.idle, _   = strconv.ParseFloat(fields[4], 64)
				ps.iowait, _ = strconv.ParseFloat(fields[5], 64)
				ps.irq, _    = strconv.ParseFloat(fields[6], 64)
				ps.sirq, _   = strconv.ParseFloat(fields[7], 64)
				ps.steal, _  = strconv.ParseFloat(fields[8], 64)
				ps.guest, _  = strconv.ParseFloat(fields[9], 64)
			}
			continue
		}

		// -----------------------------------------
		// context switches
		//    Total number of context switches across all CPUs.
		// -----------------------------------------
		if strings.HasPrefix(line, "ctxt ") {
			fields := strings.Fields(line)

			if len(fields) == 2 {
				ps.contextSwitches, _ = strconv.ParseInt(fields[1], 10, 64)
			}
			continue
		}

		// -----------------------------------------
		// processes (forks)
		//	  Number of processes and threads created, which includes (but is 
		//    not limited to) those created by calls to the fork() and clone() 
		//    system calls.
		// -----------------------------------------
		if strings.HasPrefix(line, "processes ") {
			fields := strings.Fields(line)

			if len(fields) == 2 {
				ps.forks, _ = strconv.ParseInt(fields[1], 10, 64)
				ps.vforks = 0
			}
			continue
		}	
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}

// readDentryStateStat reads /proc/sys/fs/dentry-state and returns filesystem
// dentry cache statistics used by Gonitorix for monitoring.
func readDentryStateStat() (*dentryState, error) {
	stats := &dentryState{}

	// --------------------------------------------------
	// /proc/sys/fs/dentry-state
	// --------------------------------------------------
	if data, err := os.ReadFile("/proc/sys/fs/dentry-state"); err == nil {
		fields := strings.Fields(string(data))

		if len(fields) >= 2 {
			a, _ := strconv.ParseFloat(fields[0], 64)
			b, _ := strconv.ParseFloat(fields[1], 64)

			if a+b > 0 {
				stats.dentry = (a * 100) / (a + b)
			}
		}

	} else {
		return nil, err
	}

	// --------------------------------------------------
	// /proc/sys/fs/file-nr
	// --------------------------------------------------
	if data, err := os.ReadFile("/proc/sys/fs/file-nr"); err == nil {
		fields := strings.Fields(string(data))

		if len(fields) >= 3 {
			used, _ := strconv.ParseFloat(fields[0], 64)
			max, _  := strconv.ParseFloat(fields[2], 64)

			if max > 0 {
				stats.file = (used * 100) / max
			}
		}

	} else {
		return nil, err
	}

	// --------------------------------------------------
	// /proc/sys/fs/inode-nr
	// --------------------------------------------------
	if data, err := os.ReadFile("/proc/sys/fs/inode-nr"); err == nil {
		fields := strings.Fields(string(data))

		if len(fields) >= 2 {
			a, _ := strconv.ParseFloat(fields[0], 64)
			b, _ := strconv.ParseFloat(fields[1], 64)

			if a+b > 0 {
				stats.inode = (a * 100) / (a + b)
			}
		}

	} else {
		return nil, err
	}

	return stats, nil
}