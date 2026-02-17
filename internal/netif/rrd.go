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
 
package netif

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
	step := config.NetIfCfg.Step
	heartbeat := utils.Heartbeat(step)

	for _, iface := range config.NetIfCfg.Interfaces {
		select {
			case <-ctx.Done():
				logging.Info("NETIF", "RRD creation cancelled")
				return
			default:
		}

		rrdFile := filepath.Join(
			config.GlobalCfg.RRDPath,
			config.GlobalCfg.RRDHostnamePrefix + iface.Name + ".rrd",
		)

		if _, err := os.Stat(rrdFile); err == nil {
			logging.Info("NETIF",	"RRD '%s' already exists", rrdFile,)
			continue
		}

		args := []string{
			"create", rrdFile,
			"--step", strconv.Itoa(step),

			// ----------------------------
			// Data Sources
			// ----------------------------
			fmt.Sprintf("DS:bytes_in:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:bytes_out:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:packs_in:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:packs_out:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:errors_in:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:errors_out:GAUGE:%d:0:U", heartbeat),
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

		for n := 1; n <= config.NetIfCfg.MaxHistoricYears; n++ {
			duration := n * utils.YearSeconds
			rows := utils.Rows(step, yearlyPDP, duration)

			args = append(args,
				utils.RRA("AVERAGE", 0.5, yearlyPDP, rows),
				utils.RRA("MIN", 0.5, yearlyPDP, rows),
				utils.RRA("MAX", 0.5, yearlyPDP, rows),
				utils.RRA("LAST", 0.5, yearlyPDP, rows),
			)
		}

		if err := utils.ExecCommand(ctx, "NETIF", "rrdtool", args...,); err != nil {
			logging.Error("NETIF", "Error creating RRD '%s'",	rrdFile,)
			continue
		}

		logging.Info("NETIF", "Created RRD '%s'",	rrdFile,)
	}
}

func updateRRD(ctx context.Context, rrdFile string, stats *procfs.NetIfStat) error {
	rrdata := fmt.Sprintf(
		"N:%.6f:%.6f:%.6f:%.6f:%.6f:%.6f",
		stats.RxBytes,
		stats.TxBytes,
		stats.RxPkts,
		stats.TxPkts,
		stats.RxErrors,
		stats.TxErrors,
	)

	if err := utils.ExecCommand(ctx, "NETIF", "rrdtool", "update", rrdFile, rrdata,); err != nil {
		logging.Error("NETIF", "RRDTOOL update failed for %s", rrdFile,)
		return err
	}

	return nil
}