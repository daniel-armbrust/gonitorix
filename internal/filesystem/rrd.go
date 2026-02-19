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
 
package filesystem

import (
	"context"
	"strconv"
	"os"
	"fmt"
	"strings"

	"gonitorix/internal/utils"
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
)

func createRRD(ctx context.Context) error {
	if logging.DebugEnabled() {
		logging.Debug("FILESYSTEM", "Creating filesystem RRD files")
	}

	grouped := map[string][]*filesystemDevice{}

	for _, dev := range filesystemDevices {
		grouped[dev.rrdFile] = append(grouped[dev.rrdFile], dev)
	}

	for rrdFile, devices := range grouped {
		select {
			case <-ctx.Done():
				logging.Warn("FILESYSTEM", "RRD creation cancelled by context")
				return ctx.Err()
			default:
		}

		if _, err := os.Stat(rrdFile); err == nil {
			if logging.DebugEnabled() {
				logging.Debug("FILESYSTEM", "RRD '%s' already exists", rrdFile,)
			}
			continue
		}

		step := config.FilesystemCfg.Step
		heartbeat := utils.Heartbeat(step)

		args := []string{
			"create", rrdFile,
			"--step", strconv.Itoa(step),
		}

		// ----------------------------
		// DATA SOURCES
		// ----------------------------
		for idx := range devices {
			args = append(args,
				fmt.Sprintf("DS:fs_use%d:GAUGE:%d:0:100", idx, heartbeat),
				fmt.Sprintf("DS:fs_ioa%d:GAUGE:%d:0:U", idx, heartbeat),
				fmt.Sprintf("DS:fs_tim%d:GAUGE:%d:0:U", idx, heartbeat),
				fmt.Sprintf("DS:fs_ino%d:GAUGE:%d:0:100", idx, heartbeat),
			)
		}

		// ----------------------------
		// DAILY
		// ----------------------------
		dailyRows := utils.Rows(step, 1, utils.DaySeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, 1, dailyRows),
			utils.RRA("MIN", 0.5, 1, dailyRows),
			utils.RRA("MAX", 0.5, 1, dailyRows),
			utils.RRA("LAST", 0.5, 1, dailyRows),
		)

		// ----------------------------
		// WEEKLY
		// ----------------------------
		weeklyPDP := 30
		weeklyRows := utils.Rows(step, weeklyPDP, utils.WeekSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, weeklyPDP, weeklyRows),
			utils.RRA("MIN", 0.5, weeklyPDP, weeklyRows),
			utils.RRA("MAX", 0.5, weeklyPDP, weeklyRows),
			utils.RRA("LAST", 0.5, weeklyPDP, weeklyRows),
		)

		// ----------------------------
		// MONTHLY
		// ----------------------------
		monthlyPDP := 60
		monthlyRows := utils.Rows(step, monthlyPDP, utils.MonthSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, monthlyPDP, monthlyRows),
			utils.RRA("MIN", 0.5, monthlyPDP, monthlyRows),
			utils.RRA("MAX", 0.5, monthlyPDP, monthlyRows),
			utils.RRA("LAST", 0.5, monthlyPDP, monthlyRows),
		)

		// ----------------------------
		// YEARLY
		// ----------------------------
		yearlyPDP := 1440

		for n := 1; n <= config.FilesystemCfg.MaxHistoricYears; n++ {
			duration := n * utils.YearSeconds
			rows := utils.Rows(step, yearlyPDP, duration)

			args = append(args,
				utils.RRA("AVERAGE", 0.5, yearlyPDP, rows),
				utils.RRA("MIN", 0.5, yearlyPDP, rows),
				utils.RRA("MAX", 0.5, yearlyPDP, rows),
				utils.RRA("LAST", 0.5, yearlyPDP, rows),
			)
		}

		if err := utils.ExecCommand(ctx, "FILESYSTEM", "rrdtool", args...); err != nil {
			logging.Error("FILESYSTEM",	"Failed to create RRD '%s': %v", rrdFile, err,)
			return err
		}

		logging.Info("FILESYSTEM", "Created RRD '%s' (%d filesystems)",	rrdFile, len(devices),)
	}

	return nil
}

func updateRRD(ctx context.Context, rrdFile string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	updateValue := "N:" + strings.Join(values, ":")

	args := []string{
		"update", rrdFile,
		updateValue,
	}

	if err := utils.ExecCommand(ctx, "FILESYSTEM", "rrdtool", args...); err != nil {
		logging.Error("FILESYSTEM",	"Failed to update RRD '%s': %v", rrdFile, err,)
		return err
	}

	if logging.DebugEnabled() {
		logging.Debug("FILESYSTEM", "Updated RRD '%s' with %s",	rrdFile, updateValue,)
	}

	return nil
}
