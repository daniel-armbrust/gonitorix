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

func createUptime(ctx context.Context, p *graph.GraphPeriod) {
	var defs []string
	var cdefs []string
	var draw []string

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix+
			"process-uptime-" + p.Name + ".png",
	)

	const secondsPerDay = 86400

	for i, proc := range config.ProcessCfg.Processes {
		rrdFile := filepath.Join(
			config.GlobalCfg.RRDPath,
			config.GlobalCfg.RRDHostnamePrefix+
				"process-" + utils.SanitizeName(proc.Name) + ".rrd",
		)

		alias := fmt.Sprintf("upt%d", i)
		aliasDays := fmt.Sprintf("uptd%d", i)

		defs = append(defs,
			fmt.Sprintf("DEF:%s=%s:upt:AVERAGE", alias, rrdFile),
		)

		cdefs = append(cdefs,
			fmt.Sprintf("CDEF:%s=%s,%d,/",
				aliasDays,
				alias,
				secondsPerDay,
			),
		)

		label := fmt.Sprintf("%-18s", proc.Name)

		draw = append(draw,
			fmt.Sprintf("LINE2:%s#%06X:%s",
				aliasDays,
				graph.GenerateHexColor(i),
				label,
			),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:LAST:  Cur\\: %%6.2lf d",
				aliasDays,
			),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:MIN:   Min\\: %%6.2lf d",
				aliasDays,
			),
		)

		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:MAX:   Max\\: %%6.2lf d\\l",
				aliasDays,
			),
		)
	}

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Process uptime (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: "Days",
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

	if err := utils.ExecCommand(ctx, "PROCESS", "rrdtool", args...,); err != nil {
		logging.Error("PROCESS", "Error creating uptime graph '%s': %v", graphFile, err,)
	}

	logging.Info("PROCESS", "Created uptime graph '%s'", graphFile,)
}