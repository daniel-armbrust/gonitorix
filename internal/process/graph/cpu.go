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
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/utils"
	"gonitorix/internal/graph"
)

func createCPU(ctx context.Context, p *graph.GraphPeriod) {
	var defs []string
	var cdefs []string
	var draw []string

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix+
			"process-cpu-" + p.Name + ".png",
	)

	for i, proc := range config.ProcessCfg.Processes {
		rrdFile := filepath.Join(
			config.GlobalCfg.RRDPath,
			config.GlobalCfg.RRDHostnamePrefix+
				"process-" + utils.SanitizeName(proc.Name) + ".rrd",
		)
			
		alias := fmt.Sprintf("cpu%d", i)
		aliasClean := fmt.Sprintf("%s_clean", alias)

		defs = append(defs,
			fmt.Sprintf("DEF:%s=%s:cpu:AVERAGE", alias, rrdFile),
		)

		// Remove UNKNOWN
		cdefs = append(cdefs,
			fmt.Sprintf("CDEF:%s=%s,UN,0,%s,IF",
				aliasClean,
				alias,
				alias,
			),
		)

		label := fmt.Sprintf("%-18s", proc.Name)

		draw = append(draw,
			fmt.Sprintf("LINE2:%s#%06X:%s",
				aliasClean,
				generateHexColor(i),
				label,
			),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:LAST:  Cur\\: %%6.1lf%%%%",
				aliasClean,
			),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:MIN:   Min\\: %%6.1lf%%%%",
				aliasClean,
			),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:MAX:   Max\\: %%6.1lf%%%%\\l",
				aliasClean,
			),
		)
	}

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "CPU time usage (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: "Percent (%)",
		XGrid:         p.XGrid,
		Defs:          defs,   
		CDefs:         cdefs,
		Draw:          draw,   
	}

	// Remove existing graph
	if _, err := os.Stat(graphFile); err == nil {
		_ = os.Remove(graphFile)
	}

	args := graph.BuildGraphArgs(t)

	args = append(args,
		"--upper-limit=100",
		"--lower-limit=0",
		"--rigid",
	)

	if err := utils.ExecCommand(ctx, "PROCESS", "rrdtool", args...,); err != nil {
		logging.Error("PROCESS", "Error creating image %s: %v", graphFile, err,)
	}
}
