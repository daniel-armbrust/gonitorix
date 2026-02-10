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
 
package system

import (
	"os"
	"strconv"
	"fmt"
	"context"
	
	"gonitorix/internal/config"
	"gonitorix/internal/utils"
	"gonitorix/internal/logging"
)

func createRRD(ctx context.Context) {
	rrdFile := config.GlobalCfg.RRDPath + "/" +
		       config.GlobalCfg.RRDHostnamePrefix + "system.rrd"

	step := config.SystemCfg.Step
	heartbeat := utils.Heartbeat(step)

	select {
		case <-ctx.Done():
			return
		default:
	}

	if _, err := os.Stat(rrdFile); err == nil {
		logging.Info("SYSTEM", "RRD '%s' already exists", rrdFile,)
		return
	}

	args := []string{
		"create", rrdFile,
		"--step", strconv.Itoa(step),

		// ----------------------------
		// Data Sources
		// ----------------------------
		fmt.Sprintf("DS:system_load1:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_load5:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_load15:GAUGE:%d:0:U", heartbeat),

		fmt.Sprintf("DS:system_nproc:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_npslp:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_nprun:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_npwio:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_npzom:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_npstp:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_npswp:GAUGE:%d:0:U", heartbeat),

		fmt.Sprintf("DS:system_mtotl:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_mbuff:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_mcach:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_mfree:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_macti:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_minac:GAUGE:%d:0:U", heartbeat),

		fmt.Sprintf("DS:system_entrop:GAUGE:%d:0:U", heartbeat),
		fmt.Sprintf("DS:system_uptime:GAUGE:%d:0:U", heartbeat),
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

	for n := 1; n <= config.SystemCfg.MaxHistoricYears; n++ {
		duration := n * utils.YearSeconds
		rows := utils.Rows(step, yearlyPDP, duration)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, yearlyPDP, rows),
			utils.RRA("MIN", 0.5, yearlyPDP, rows),
			utils.RRA("MAX", 0.5, yearlyPDP, rows),
			utils.RRA("LAST", 0.5, yearlyPDP, rows),
		)
	}

	if err := utils.ExecCommand(ctx, "SYSTEM", "rrdtool", args...,); err != nil {
		logging.Error("SYSTEM", "Error creating RRD '%s'", rrdFile,)
		return
	}

	logging.Info("SYSTEM", "Created RRD '%s'", rrdFile,)
}

func updateRRD(ctx context.Context) error {
	rrdFile := config.GlobalCfg.RRDPath + "/" +
		       config.GlobalCfg.RRDHostnamePrefix + "system.rrd"

	memory, err := readMemory(ctx)
	if err != nil {
		return err
	}

	loadavg, err := readLoadAvg(ctx)
	if err != nil {
		return err
	}

	entropy, err := readEntropy(ctx)
	if err != nil {
		return err
	}

	procinfo, err := readProcInfo(ctx)
	if err != nil {
		return err
	}

	uptime, err := readUptime(ctx)
	if err != nil {
		return err
	}

	rrdata := fmt.Sprintf(
		"N:%.2f:%.2f:%.2f:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%.0f",

		// load
		loadavg["load1"],
		loadavg["load5"],
		loadavg["load15"],

		// processes
		procinfo["total"],
		procinfo["sleep"],
		procinfo["run"],
		procinfo["wio"],
		procinfo["zombie"],
		procinfo["stop"],
		procinfo["swap"],

		// memory
		memory["MemTotal"],
		memory["Buffers"],
		memory["Cached"],
		memory["MemFree"],
		memory["Active"],
		memory["Inactive"],

		// other
		entropy,
		uptime,
	)

	if err := utils.ExecCommand(ctx, "SYSTEM", "rrdtool", "update", rrdFile, rrdata,); err != nil {
		logging.Error("SYSTEM", "RRDTOOL update failed for %s", rrdFile,)
		return err
	}

	return nil
}
