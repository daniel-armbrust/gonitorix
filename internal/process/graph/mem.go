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

func createMem(ctx context.Context, p *graph.GraphPeriod) {
	var defs []string
	var cdefs []string
	var draw []string

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix+
			"process-mem-" + p.Name + ".png",
	)

	for i, proc := range config.ProcessCfg.Processes {
		rrdFile := filepath.Join(
			config.GlobalCfg.RRDPath,
			config.GlobalCfg.RRDHostnamePrefix+
				"process-" + utils.SanitizeName(proc.Name) + ".rrd",
		)

		alias := fmt.Sprintf("mem%d", i)
		aliasMB := fmt.Sprintf("%s_mb", alias)

		// -------------------------------------------------
		// DEF (memory stored in bytes)
		// -------------------------------------------------
		defs = append(defs,
			fmt.Sprintf("DEF:%s=%s:mem:AVERAGE", alias, rrdFile),
		)

		// -------------------------------------------------
		// Convert bytes -> MB (legend only)
		// -------------------------------------------------
		cdefs = append(cdefs,
			fmt.Sprintf("CDEF:%s=%s,1048576,/",	aliasMB, alias,),
		)

		// -------------------------------------------------
		// Draw line
		// -------------------------------------------------
		label := fmt.Sprintf("%-18s", proc.Name)

		draw = append(draw, 
			fmt.Sprintf("LINE2:%s#%06X:%s", alias, graph.GenerateHexColor(i), label,),
		)

		// -------------------------------------------------
		// Legend (Monitorix style)
		// -------------------------------------------------
		draw = append(draw, 
			fmt.Sprintf("GPRINT:%s:LAST:  Cur\\: %%6.0lfM", aliasMB,),
        )

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:MIN:   Min\\: %%6.0lfM", aliasMB,),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:MAX:   Max\\: %%6.0lfM\\l", aliasMB,),
		)
	}

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Memory usage (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: "bytes",
		XGrid:         p.XGrid,
		Defs:          defs,
		CDefs:         cdefs,
		Draw:          draw,
	}

	// Remove existing image if present
	if _, err := os.Stat(graphFile); err == nil {
		_ = os.Remove(graphFile)
	}

	args := graph.BuildGraphArgs(t)

	if err := utils.ExecCommand(ctx, "PROCESS", "rrdtool", args...,); err != nil {
		logging.Error("PROCESS", "Error creating memory graph '%s': %v", graphFile, err,)
	}

	logging.Info("PROCESS", "Created memory graph '%s'", graphFile,)
}