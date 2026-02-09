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

package latency

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"fmt"

	"gonitorix/internal/config"
	"gonitorix/internal/utils"
)

func createRRD() {
	for _, host := range config.LatencyCfg.Hosts {
		rrdFile := config.GlobalCfg.RRDPath + "/" + host.RRDFile

		step := config.LatencyCfg.Step
		heartbeat := utils.Heartbeat(step)

		_, err := os.Stat(rrdFile)

		if os.IsNotExist(err) {
			// https://github.com/sandromarcell/rrd-rttping
			args := []string{
				"create", rrdFile,
				"--step", strconv.Itoa(step),

				// --------------------------------------------------
				// Data Sources
				// --------------------------------------------------
				fmt.Sprintf("DS:min:GAUGE:%d:0:U",  heartbeat),
				fmt.Sprintf("DS:avg:GAUGE:%d:0:U",  heartbeat),
				fmt.Sprintf("DS:max:GAUGE:%d:0:U",  heartbeat),
				fmt.Sprintf("DS:loss:GAUGE:%d:0:U", heartbeat),
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
}

func updateRRD(rrdFile string, data *pingResult) {
	rrdFile = config.GlobalCfg.RRDPath + "/" + rrdFile
	
	rrdata := fmt.Sprintf(
		"N:%s:%s:%s:%s",

		// Latency (ms)
		utils.RRDfloat(data.min, 2),
		utils.RRDfloat(data.avg, 2),
		utils.RRDfloat(data.max, 2),

		// Packet loss (%)
		utils.RRDfloat(data.loss, 2),
	)

	cmd := exec.Command(
		"rrdtool",
		"update",
		rrdFile,
		rrdata,
	)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf(
			"RRDTOOL update failed for %s: %v | output: %s\n",
			rrdFile,
			err,
			string(out),
		)
	}
}