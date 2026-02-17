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

func createContextSwitches(ctx context.Context, p *graph.GraphPeriod) {
	var defs []string
	var cdefs []string
	var draw []string

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix+
			"process-ctxswitches-" + p.Name + ".png",
	)

	for i, proc := range config.ProcessCfg.Processes {
		rrdFile := filepath.Join(
			config.GlobalCfg.RRDPath,
			config.GlobalCfg.RRDHostnamePrefix+
				"process-"+utils.SanitizeName(proc.Name)+".rrd",
		)

		vcs := fmt.Sprintf("vcs%d", i)
		ics := fmt.Sprintf("ics%d", i)
		nics := fmt.Sprintf("n_ics%d", i)

		defs = append(defs,
			fmt.Sprintf("DEF:%s=%s:vcs:AVERAGE", vcs, rrdFile),
			fmt.Sprintf("DEF:%s=%s:ics:AVERAGE", ics, rrdFile),
		)

		// Invert ICS
		cdefs = append(cdefs,
			fmt.Sprintf("CDEF:%s=%s,-1,*", nics, ics),
		)

		label := fmt.Sprintf("%-18s", proc.Name)

		// Draw voluntary (positive)
		draw = append(draw,
			fmt.Sprintf("AREA:%s#%06X:%s",
				vcs,
				generateHexColor(i),
				label,
			),
		)

		// Draw involuntary (negative)
		draw = append(draw,
			fmt.Sprintf("AREA:%s#%06X",
				nics,
				generateHexColor(i+20), 
			),
		)

		total := fmt.Sprintf("tcs%d", i)

		cdefs = append(cdefs,
			fmt.Sprintf("CDEF:%s=%s,%s,+", total, vcs, ics),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:LAST:  Cur\\: %%6.0lf",
				total,
			),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:MIN:   Min\\: %%6.0lf",
				total,
			),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:MAX:   Max\\: %%6.0lf\\l",
				total,
			),
		)
	}

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Context switches (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: "Nonvoluntary + voluntary/s",
		XGrid:         p.XGrid,
		Defs:          defs,
		CDefs:         cdefs,
		Draw:          draw,
	}

	if _, err := os.Stat(graphFile); err == nil {
		_ = os.Remove(graphFile)
	}

	args := graph.BuildGraphArgs(t)

	if err := utils.ExecCommand(ctx, "PROCESS", "rrdtool", args...,); err != nil {
		logging.Error("PROCESS", "Error creating context switch image %s: %v", graphFile, err,)
	}
}