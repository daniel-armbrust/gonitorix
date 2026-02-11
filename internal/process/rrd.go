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
	"os"
	"fmt"
	"context"
	"strconv"
	
	"gonitorix/internal/config"
	"gonitorix/internal/utils"
	"gonitorix/internal/logging"
)

// buildRRDFileList generates the list of RRD file paths required to store
// process metrics based on the number of configured processes and the
// maximum allowed processes per RRD file.
func buildRRDFileList() []string {
	var rrdFileList []string

	numProcs := len(config.ProcessCfg.Processes)

	numRRDs := (numProcs + MaxProcPerRRD - 1) / MaxProcPerRRD

	for i := 1; i <= numRRDs; i++ {
		rrdFile := config.GlobalCfg.RRDPath + "/" +
			config.GlobalCfg.RRDHostnamePrefix +
			"process-" + strconv.Itoa(i) + ".rrd"

		rrdFileList = append(rrdFileList, rrdFile)
	}

	return rrdFileList
}

// buildRRDArgs builds the list of rrdtool create arguments (DS and RRA)
// for a single RRD file, based on the RRD index, the number of processes
// stored in that file, and the calculated heartbeat interval.
func buildRRDArgs(startProcIndex int, procCount int, step int, heartbeat int,) []string {
	var args []string

	metrics := []struct {
		Name string
		Max  string
	}{
		{"cpu", "100"},
		{"mem", "U"},
		{"dsk", "U"},
		{"net", "U"},
		{"nof", "U"},
		{"pro", "U"},
		{"nth", "U"},
		{"vcs", "U"},
		{"ics", "U"},
		{"upt", "U"},
		{"va2", "U"},
	}

	for p := 0; p < procCount; p++ {
		globalIdx := startProcIndex + p

		for _, m := range metrics {
			ds := fmt.Sprintf("DS:proc%d_%s:GAUGE:%d:0:%s",	globalIdx, m.Name, heartbeat, m.Max,)
			args = append(args, ds)
		}
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

	return args
}

func createRRD(ctx context.Context) {
	rrdFileList := buildRRDFileList()

	numProcs := len(config.ProcessCfg.Processes)

	step := config.ProcessCfg.Step
	heartbeat := utils.Heartbeat(step)

	for i := 0; i < len(rrdFileList); i++ {
		rrdFile := rrdFileList[i]

		if _, err := os.Stat(rrdFile); err == nil {
			logging.Info("PROCESS", "RRD '%s' already exists", rrdFile)
			continue
		}

		start := i * MaxProcPerRRD
		end := start + MaxProcPerRRD

		if end > numProcs {
			end = numProcs
		}

		procsInThisRRD := end - start

		if procsInThisRRD <= 0 {
			continue
		}

		args := buildRRDArgs(i, procsInThisRRD, step, heartbeat,)

		cmdArgs := []string{"create", rrdFile, "--step=" + strconv.Itoa(step),}
		cmdArgs = append(cmdArgs, args...)

		if err := utils.ExecCommand(ctx, "PROCESS", "rrdtool", cmdArgs...); err != nil {
			logging.Error("PROCESS", "Error creating RRD '%s'", rrdFile)
			continue
		}

		logging.Info("PROCESS", "RRD '%s' created successfully", rrdFile)
	}
}
