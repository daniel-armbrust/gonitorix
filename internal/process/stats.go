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
)

// FindProcessPIDs scans the process table and returns all PIDs whose
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

