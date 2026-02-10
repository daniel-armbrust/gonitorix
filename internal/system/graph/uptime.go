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
 
package graph

import (
	"fmt"
	"os"
	"strings"
	"context"

	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/graph"
	"gonitorix/internal/utils"
)

type uptimeUnit struct {
	yTitle string
	unit   int
	format string
}

func uptimeUnitConfig(timeUnit string) uptimeUnit {
	switch strings.ToLower(timeUnit) {

	case "minute":
		return uptimeUnit{
			yTitle: "Minutes",
			unit:   60,
			format: "%5.0lf",
		}

	case "hour":
		return uptimeUnit{
			yTitle: "Hours",
			unit:   3600,
			format: "%5.0lf",
		}

	default:
		return uptimeUnit{
			yTitle: "Days",
			unit:   86400,
			format: "%5.1lf",
		}
	}
}

// createUptime generates RRD graphs showing system uptime for the given
// graph period.
func createUptime(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := config.GlobalCfg.RRDPath + "/" +
		       config.GlobalCfg.RRDHostnamePrefix + "system.rrd"

	graphFile := config.GlobalCfg.GraphPath + "/" +
		         config.GlobalCfg.RRDHostnamePrefix +
		         "uptime-" + p.Name + ".png"

	u := uptimeUnitConfig("")

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Uptime (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: u.yTitle,
		Width:         450,
		Height:        150,
		XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:uptime=%s:system_uptime:AVERAGE", rrdFile),
		},

		CDefs: []string{
			fmt.Sprintf("CDEF:uptime_days=uptime,%d,/", u.unit),
			"CDEF:allvalues=uptime",
		},

		Draw: []string{
			"LINE2:uptime_days#EE44EE:Uptime",
			fmt.Sprintf(
				"GPRINT:uptime_days:LAST: Current\\:%s\\n",
				u.format,
			),
		},
	}

	// Remove the PNG file if it already exists.
	if _, err := os.Stat(graphFile); err == nil {
		if err := os.Remove(graphFile); err != nil {
			logging.Warn("SYSTEM", "Failed to remove existing graph %s: %v", graphFile,	err,)
		}
	}

	args := graph.BuildGraphArgs(t)

	if err := utils.ExecCommand(ctx, "SYSTEM", "rrdtool", args...,); err != nil {
		logging.Error("SYSTEM",	"Error creating image %s", graphFile,)
	}
}
