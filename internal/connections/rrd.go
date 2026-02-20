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

 package connections

 import (
	"os"
	"strconv"
	"fmt"
	"context"
	"path/filepath"
	
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/utils"
)

func createRRD(ctx context.Context) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "connections.rrd",
	)

	step := config.ConnectionsCfg.Step
	heartbeat := utils.Heartbeat(step)

	_, err := os.Stat(rrdFile)

	if os.IsNotExist(err) {
		args := []string{
			"create", rrdFile,
			"--step", strconv.Itoa(step),

			// --------------------------------------------------
			// IPv4 Data Sources
			// --------------------------------------------------
			fmt.Sprintf("DS:nstat4_closed:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_listen:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_synSent:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_synRecv:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_estblshd:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_finWait1:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_finWait2:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_closing:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_timeWait:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_closeWait:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_lastAck:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_unknown:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_udp:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_val1:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_val2:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_val3:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_val4:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat4_val5:GAUGE:%d:0:U", heartbeat),

			// --------------------------------------------------
			// IPv6 Data Sources
			// --------------------------------------------------
			fmt.Sprintf("DS:nstat6_closed:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_listen:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_synSent:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_synRecv:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_estblshd:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_finWait1:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_finWait2:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_closing:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_timeWait:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_closeWait:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_lastAck:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_unknown:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_udp:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_val1:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_val2:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_val3:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_val4:GAUGE:%d:0:U", heartbeat),
			fmt.Sprintf("DS:nstat6_val5:GAUGE:%d:0:U", heartbeat),
		}

		// --------------------------------------------------
		// DAILY
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

		if err := utils.ExecCommand(ctx, "CONNECTIONS", "rrdtool", args...); err != nil {
			logging.Error("CONNECTIONS", "Error creating RRD '%s'", rrdFile)
			return
		}

		logging.Info("CONNECTIONS", "Created RRD '%s'", rrdFile)

	} else {
		logging.Info("CONNECTIONS", "RRD '%s' already exists", rrdFile)
	}
}

func updateRRD(ctx context.Context, ipv4, ipv6 connStats) error {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "connections.rrd",
	)

	value := fmt.Sprintf(
		"N:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:"+
			"%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d",
		ipv4.closed,
		ipv4.listen,
		ipv4.synSent,
		ipv4.synRecv,
		ipv4.estab,
		ipv4.finWait1,
		ipv4.finWait2,
		ipv4.closing,
		ipv4.timeWait,
		ipv4.closeWait,
		ipv4.lastAck,
		ipv4.unknown,
		ipv4.udp,
		0, 0, 0, 0, 0, // val1–val5 IPv4

		ipv6.closed,
		ipv6.listen,
		ipv6.synSent,
		ipv6.synRecv,
		ipv6.estab,
		ipv6.finWait1,
		ipv6.finWait2,
		ipv6.closing,
		ipv6.timeWait,
		ipv6.closeWait,
		ipv6.lastAck,
		ipv6.unknown,
		ipv6.udp,
		0, 0, 0, 0, 0, // val1–val5 IPv6
	)

	args := []string{
		"update", rrdFile, value,
	}

	if err := utils.ExecCommand(ctx, "CONNECTIONS", "rrdtool", args...); err != nil {
		logging.Error("CONNECTIONS", "Error updating RRD '%s'", rrdFile)
		return err
	}

	return nil
}