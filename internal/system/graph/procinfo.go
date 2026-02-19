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
	"gonitorix/internal/graph"
	"gonitorix/internal/logging"	
	"gonitorix/internal/utils"
)

// createProcInfo generates RRD graphs showing process state distribution
// for the given graph period.
func createProcInfo(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "system.rrd",
	)

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix + "proc-" + p.Name + ".png",
	)

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Active Processes (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: "Processes",
		XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:nproc=%s:system_nproc:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npslp=%s:system_npslp:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:nprun=%s:system_nprun:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npwio=%s:system_npwio:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npzom=%s:system_npzom:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npstp=%s:system_npstp:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npswp=%s:system_npswp:AVERAGE", rrdFile),
		},

		CDefs: []string{
			"CDEF:allvalues=nproc,npslp,nprun,npwio,npzom,npstp,npswp,+,+,+,+,+,+",
		},

		Draw: []string{
			"AREA:npslp#33CC33:Sleeping         ",
			"GPRINT:npslp:LAST:Cur\\:%5.0lf",
			"GPRINT:npslp:MIN:Min\\:%5.0lf",
			"GPRINT:npslp:MAX:Max\\:%5.0lf\\l",

			"LINE1:npwio#FFCC00:Wait I/O         ",
			"GPRINT:npwio:LAST:Cur\\:%5.0lf",
			"GPRINT:npwio:MIN:Min\\:%5.0lf",
			"GPRINT:npwio:MAX:Max\\:%5.0lf\\l",

			"LINE1:npzom#AA00FF:Zombie           ",
			"GPRINT:npzom:LAST:Cur\\:%5.0lf",
			"GPRINT:npzom:MIN:Min\\:%5.0lf",
			"GPRINT:npzom:MAX:Max\\:%5.0lf\\l",

			"LINE1:npstp#00AAAA:Stopped          ",
			"GPRINT:npstp:LAST:Cur\\:%5.0lf",
			"GPRINT:npstp:MIN:Min\\:%5.0lf",
			"GPRINT:npstp:MAX:Max\\:%5.0lf\\l",

			"LINE1:nprun#FF0000:Running          ",
			"GPRINT:nprun:LAST:Cur\\:%5.0lf",
			"GPRINT:nprun:MIN:Min\\:%5.0lf",
			"GPRINT:nprun:MAX:Max\\:%5.0lf\\l",

			"LINE2:nproc#FFFFFF:Total Processes  ",
			"GPRINT:nproc:LAST:Cur\\:%5.0lf",
			"GPRINT:nproc:MIN:Min\\:%5.0lf",
			"GPRINT:nproc:MAX:Max\\:%5.0lf\\l",
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
		logging.Error("SYSTEM",	"Failed to create system process states graph %s: %v",	graphFile, err,)
	}

	logging.Info("SYSTEM", "Created system process states graph '%s'", graphFile,)
}
