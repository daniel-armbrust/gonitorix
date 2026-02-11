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
	"fmt"
	"context"
	
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/utils"
)

func createRRD(ctx context.Context) {
	rrdFile := config.GlobalCfg.RRDPath + "/" + 
	           config.GlobalCfg.RRDHostnamePrefix + "kernel.rrd"

	step := config.KernelCfg.Step
	heartbeat := utils.Heartbeat(step)

	_, err := os.Stat(rrdFile)

	if os.IsNotExist(err) {		
		args := []string{
			"create", rrdFile,
			"--step", strconv.Itoa(step),

			// --------------------------------------------------
			// Data Sources
			// --------------------------------------------------
			fmt.Sprintf("DS:kern_user:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_nice:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_sys:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_idle:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_iow:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_irq:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_sirq:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_steal:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_guest:GAUGE:%d:0:100", heartbeat),
			
			fmt.Sprintf("DS:kern_cs:COUNTER:%d:0:U", heartbeat),
			fmt.Sprintf("DS:kern_forks:COUNTER:%d:0:U", heartbeat),
			fmt.Sprintf("DS:kern_vforks:COUNTER:%d:0:U", heartbeat),

			fmt.Sprintf("DS:kern_dentry:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_file:GAUGE:%d:0:100", heartbeat),
			fmt.Sprintf("DS:kern_inode:GAUGE:%d:0:100", heartbeat),			

			// fmt.Sprintf("DS:kern_val03:GAUGE:%d:0:100", heartbeat),
			// fmt.Sprintf("DS:kern_val04:GAUGE:%d:0:100", heartbeat),
			// fmt.Sprintf("DS:kern_val05:GAUGE:%d:0:100", heartbeat),
		}

		// --------------------------------------------------
		// DAILY
		// --------------------------------------------------
		dailyRows := utils.Rows(step, 1, utils.DaySeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, 1, dailyRows),
		)

		// --------------------------------------------------
		// WEEKLY
		// --------------------------------------------------
		weeklyPDP := 30
		weeklyRows := utils.Rows(step, weeklyPDP, utils.WeekSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, weeklyPDP, weeklyRows),
		)

		// --------------------------------------------------
		// MONTHLY
		// --------------------------------------------------
		monthlyPDP := 60
		monthlyRows := utils.Rows(step, monthlyPDP, utils.MonthSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, monthlyPDP, monthlyRows),
		)

		// --------------------------------------------------
		// YEARLY
		// --------------------------------------------------
		yearlyPDP := 1440

		rows := utils.Rows(step, yearlyPDP, utils.YearSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, yearlyPDP, rows),
		)

		if err := utils.ExecCommand(ctx, "KERNEL", "rrdtool", args...); err != nil {
			logging.Error("KERNEL", "Error creating RRD '%s'", rrdFile)
			return
		}

		logging.Info("KERNEL", "Created RRD '%s'", rrdFile)
	} else {
		logging.Info("KERNEL", "RRD '%s' already exists", rrdFile,)
	}
}

func updateRRD(ctx context.Context, stats *procStatDentryStat) error {
	rrdFile := config.GlobalCfg.RRDPath + "/" +
		       config.GlobalCfg.RRDHostnamePrefix + "kernel.rrd"

	rrdata := fmt.Sprintf(
		"N:%s:%s:%s:%s:%s:%s:%s:%s:%s:%d:%d:%d:%s:%s:%s",

		// CPU %
		utils.RRDfloat(stats.user, 6),
		utils.RRDfloat(stats.nice, 6),
		utils.RRDfloat(stats.sys, 6),
		utils.RRDfloat(stats.idle, 6),
		utils.RRDfloat(stats.iowait, 6),
		utils.RRDfloat(stats.irq, 6),
		utils.RRDfloat(stats.sirq, 6),
		utils.RRDfloat(stats.steal, 6),
		utils.RRDfloat(stats.guest, 6),

		// Counters
		stats.contextSwitches,
		stats.forks,
		stats.vforks,

		// VFS %
		utils.RRDfloat(stats.dentry, 2),
		utils.RRDfloat(stats.file, 2),
		utils.RRDfloat(stats.inode, 2),
	)

	if err := utils.ExecCommand(ctx, "KERNEL", "rrdtool", "update", rrdFile, rrdata,); err != nil {
		logging.Error("KERNEL", "RRDTOOL update failed for %s", rrdFile,)

		return err
	}

	return nil
}
