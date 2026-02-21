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

 package interrupts

 import (
	"os"
	"strconv"
	"fmt"
	"context"
	"path/filepath"
	
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/utils"
	"gonitorix/internal/procfs"
)

func createRRD(ctx context.Context) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "interrupts.rrd",
	)

	step := config.InterruptsCfg.Step
	heartbeat := utils.Heartbeat(step)

	_, err := os.Stat(rrdFile)

	if os.IsNotExist(err) {
		args := []string{
			"create", rrdFile,
			"--step", strconv.Itoa(step),

			// --------------------------------------------------
			// Interrupt Total (from /proc/stat intr)
			// --------------------------------------------------
			fmt.Sprintf("DS:intr_total:COUNTER:%d:0:U", heartbeat),
		}

		// --------------------------------------------------
		// DAILY (high resolution)
		// --------------------------------------------------
		dailyRows := utils.Rows(step, 1, utils.DaySeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, 1, dailyRows),
			utils.RRA("MIN",     0.5, 1, dailyRows),
			utils.RRA("MAX",     0.5, 1, dailyRows),
			utils.RRA("LAST",    0.5, 1, dailyRows),
		)

		// --------------------------------------------------
		// WEEKLY
		// --------------------------------------------------
		weeklyPDP := 30
		weeklyRows := utils.Rows(step, weeklyPDP, utils.WeekSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, weeklyPDP, weeklyRows),
			utils.RRA("MIN",     0.5, weeklyPDP, weeklyRows),
			utils.RRA("MAX",     0.5, weeklyPDP, weeklyRows),
			utils.RRA("LAST",    0.5, weeklyPDP, weeklyRows),
		)

		// --------------------------------------------------
		// MONTHLY
		// --------------------------------------------------
		monthlyPDP := 60
		monthlyRows := utils.Rows(step, monthlyPDP, utils.MonthSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, monthlyPDP, monthlyRows),
			utils.RRA("MIN",     0.5, monthlyPDP, monthlyRows),
			utils.RRA("MAX",     0.5, monthlyPDP, monthlyRows),
			utils.RRA("LAST",    0.5, monthlyPDP, monthlyRows),
		)

		// --------------------------------------------------
		// YEARLY
		// --------------------------------------------------
		yearlyPDP := 1440
		yearlyRows := utils.Rows(step, yearlyPDP, utils.YearSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, yearlyPDP, yearlyRows),
			utils.RRA("MIN",     0.5, yearlyPDP, yearlyRows),
			utils.RRA("MAX",     0.5, yearlyPDP, yearlyRows),
			utils.RRA("LAST",    0.5, yearlyPDP, yearlyRows),
		)

		if err := utils.ExecCommand(ctx, "INTERRUPTS", "rrdtool", args...); err != nil {
			logging.Error("INTERRUPTS", "Error creating RRD '%s'", rrdFile)
			return
		}

		logging.Info("INTERRUPTS", "Created RRD '%s'", rrdFile)

	} else {
		logging.Info("INTERRUPTS", "RRD '%s' already exists", rrdFile)
	}
}

func updateRRD(ctx context.Context, stats *procfs.InterruptStat) error {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "interrupts.rrd",
	)

	// Update the RRD with the cumulative total IRQ counter 
	// (COUNTER DS handles rate calculation).
	value := "N:" + strconv.FormatUint(stats.Total, 10)

	args := []string{
		"update", rrdFile, value,
	}

	if err := utils.ExecCommand(ctx, "INTERRUPTS", "rrdtool", args...); err != nil {
		logging.Error("INTERRUPTS", "Error updating RRD '%s'", rrdFile)
		return err
	}

	return nil
}