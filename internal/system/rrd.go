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
	"os/exec"
	"strconv"
	"log"
	"fmt"
	
	"gonitorix/internal/config"
	"gonitorix/internal/utils"
)

func createRRD() {
	rrdPath := config.GlobalCfg.RRDPath
	rrdFile := rrdPath + "/system.rrd"

	step := config.NetIfCfg.Step
	heartbeat := utils.Heartbeat(step)

	_, err := os.Stat(rrdFile)

	if os.IsNotExist(err) {
		args := []string{
			"create", rrdFile,
			"--step", strconv.Itoa(step),

			// --------------------------------------------------
			// Data Sources
			// --------------------------------------------------
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

		// --------------------------------------------------
		// DAILY
		// --------------------------------------------------
		dailyRows := utils.Rows(step, 1, utils.DaySeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, 1, dailyRows),
			utils.RRA("MIN", 0.5, 1, dailyRows),
			utils.RRA("MAX", 0.5, 1, dailyRows),
			utils.RRA("LAST", 0.5, 1, dailyRows),
		)

		// --------------------------------------------------
		// WEEKLY
		// --------------------------------------------------
		weeklyPDP := 30
		weeklyRows := utils.Rows(step, weeklyPDP, utils.WeekSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, weeklyPDP, weeklyRows),
			utils.RRA("MIN", 0.5, weeklyPDP, weeklyRows),
			utils.RRA("MAX", 0.5, weeklyPDP, weeklyRows),
			utils.RRA("LAST", 0.5, weeklyPDP, weeklyRows),
		)

		// --------------------------------------------------
		// MONTHLY
		// --------------------------------------------------
		monthlyPDP := 60
		monthlyRows := utils.Rows(step, monthlyPDP, utils.MonthSeconds)

		args = append(args,
			utils.RRA("AVERAGE", 0.5, monthlyPDP, monthlyRows),
			utils.RRA("MIN", 0.5, monthlyPDP, monthlyRows),
			utils.RRA("MAX", 0.5, monthlyPDP, monthlyRows),
			utils.RRA("LAST", 0.5, monthlyPDP, monthlyRows),
		)

		// --------------------------------------------------
		// YEARLY
		// --------------------------------------------------
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

		cmd := exec.Command("rrdtool", args...)			
		_, err := cmd.CombinedOutput()

		if err != nil {
			log.Printf("Error creating RRD '%s': %v\n", rrdFile, err)
			return
		}

		log.Printf("Creating RRD '%s'", rrdFile)	
	} else {
		log.Printf("RRD '%s' already exists", rrdFile)
	}	
}

func updateRRD() {
	rrdPath := config.GlobalCfg.RRDPath
	rrdFile := rrdPath + "/system.rrd"

	memory, err := readMemory()

	if err != nil {
		log.Printf("readMemory failed: %w\n", err)
	}

	loadavg, err := readLoadAvg()

	if err != nil {
		log.Printf("readLoadAvg failed: %w\n", err)
	}

	entropy, err := readEntropy()

	if err != nil {
		log.Printf("readEntropy failed: %w\n", err)
	}

	procinfo, err := readProcInfo()

	if err != nil {
		log.Printf("readProcInfo failed: %w\n", err)
	}

	uptime, err := readUptime()

	if err != nil {
		log.Printf("readUptime failed: %w\n", err)
	}

	rrdata := fmt.Sprintf(
		"N:%.2f:%.2f:%.2f:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%.0f",

		// load
		loadavg["load1"],
		loadavg["load5"],
		loadavg["load15"],

		// processos
		procinfo["total"],
		procinfo["sleep"],
		procinfo["run"],
		procinfo["wio"],
		procinfo["zombie"],
		procinfo["stop"],
		procinfo["swap"],

		// memória
		memory["MemTotal"],
		memory["Buffers"],
		memory["Cached"],
		memory["MemFree"],
		memory["Active"],
		memory["Inactive"],

		// outros
		entropy,
		uptime,
	)

	cmd := exec.Command(
		"rrdtool", "update", rrdFile, rrdata,
	)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("RRDTOOL update failed: %v | output: %s\n", err, out)
	}
}