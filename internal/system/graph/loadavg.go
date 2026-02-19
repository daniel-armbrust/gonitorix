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
	"context"
	"path/filepath"
		
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/graph"
	"gonitorix/internal/utils"
)

// createLoadavg generates RRD graphs showing system load averages
// for the given graph period.
func createLoadavg(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "system.rrd",
	)

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix + "loadavg-" + p.Name + ".png",
	)
	
	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "System Load (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: "Load average",
		XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:load1=%s:system_load1:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:load5=%s:system_load5:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:load15=%s:system_load15:AVERAGE", rrdFile),
		},

		Draw: []string{
			// 1 min
			fmt.Sprintf("LINE2:load1#4444EE:%-18s", "1 min average"),
			"GPRINT:load1:LAST:  Cur\\: %6.2lf",
			"GPRINT:load1:AVERAGE:  Avg\\: %6.2lf",
			"GPRINT:load1:MIN:  Min\\: %6.2lf",
			"GPRINT:load1:MAX:  Max\\: %6.2lf\\l",

			// 5 min
			fmt.Sprintf("LINE2:load5#EEEE00:%-18s", "5 min average"),
			"GPRINT:load5:LAST:  Cur\\: %6.2lf",
			"GPRINT:load5:AVERAGE:  Avg\\: %6.2lf",
			"GPRINT:load5:MIN:  Min\\: %6.2lf",
			"GPRINT:load5:MAX:  Max\\: %6.2lf\\l",

			// 15 min
			fmt.Sprintf("LINE2:load15#00EEEE:%-18s", "15 min average"),
			"GPRINT:load15:LAST:  Cur\\: %6.2lf",
			"GPRINT:load15:AVERAGE:  Avg\\: %6.2lf",
			"GPRINT:load15:MIN:  Min\\: %6.2lf",
			"GPRINT:load15:MAX:  Max\\: %6.2lf\\l",
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
		logging.Error("SYSTEM",	"Failed to create system load average graph '%s': %v", graphFile, err,)
	}

	logging.Info("SYSTEM", "Created system load average graph '%s'", graphFile,)
}
