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
	"context"
	"path/filepath"
				
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/utils"
	"gonitorix/internal/graph"
)

func createConnUDPStats(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix+"connections.rrd",
	)

	var defs []string
	var draw []string

	states := []struct {
		ds    string
		label string
	}{
		{"nstat4_udp", "UDP v4"},
		{"nstat6_udp", "UDP v6"},
	}

	for i, state := range states {

		alias := fmt.Sprintf("udp%d", i)

		// -----------------------------------------
		// DEF
		// -----------------------------------------
		defs = append(defs,
			fmt.Sprintf(
				"DEF:%s=%s:%s:AVERAGE",
				alias,
				rrdFile,
				state.ds,
			),
		)

		// -----------------------------------------
		// LINE
		// -----------------------------------------
		draw = append(draw,
			fmt.Sprintf(
				"LINE2:%s#%06X:%-12s",
				alias,
				graph.GenerateHexColor(i),
				state.label,
			),
		)

		// -----------------------------------------
		// GPRINT
		// -----------------------------------------
		draw = append(draw,
			fmt.Sprintf("GPRINT:%s:LAST:  Cur\\: %%6.0lf", alias),
			fmt.Sprintf("GPRINT:%s:MIN:   Min\\: %%6.0lf", alias),
			fmt.Sprintf("GPRINT:%s:MAX:   Max\\: %%6.0lf\\l", alias),
		)
	}

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		fmt.Sprintf(
			"%sconnections-udp-%s.png",
			config.GlobalCfg.RRDHostnamePrefix,
			p.Name,
		),
	)

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         fmt.Sprintf("UDP Listening Sockets (%s)", p.Name),
		Start:         p.Start,
		VerticalLabel: "Listen",
		XGrid:         p.XGrid,
		Defs:          defs,
		Draw:          draw,
	}

	args := graph.BuildGraphArgs(t)

	if err := utils.ExecCommand(ctx, "CONNECTIONS", "rrdtool", args...); err != nil {
		logging.Error("CONNECTIONS", "Failed to create UDP graph '%s': %v", graphFile, err,)
		return
	}

	logging.Info("CONNECTIONS", "Created UDP graph '%s'", graphFile)
}