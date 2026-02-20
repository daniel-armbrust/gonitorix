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

// createConnectionsPassiveClose generates a graph showing TCP states related
// to Passive Close operations.
//
// Passive Close occurs when the remote peer initiates the connection
// termination by sending the first FIN packet. The local host responds
// and transitions through states such as:
//
//   - CLOSE_WAIT
//   - LAST_ACK
//
// CLOSE_WAIT is particularly important from a monitoring perspective.
// A high number of CLOSE_WAIT connections may indicate that the application
// is not properly closing sockets, potentially leading to resource leaks.
//
// This graph includes both IPv4 and IPv6 states to provide full visibility
// into passive shutdown behavior across protocols.
//
// The implementation follows the standard Gonitorix graph pattern:
//   - Defines RRD data sources (DEF)
//   - Draws each state as a LINE2
//   - Prints LAST, MIN, and MAX values
func createConnPassiveClose(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "connections.rrd",
	)

	var defs []string
	var draw []string

	states := []struct {
		ds    string
		label string
	}{
		{"nstat4_closeWait", "CLOSE_WAIT v4"},
		{"nstat6_closeWait", "CLOSE_WAIT v6"},
		{"nstat4_lastAck", "LAST_ACK v4"},
		{"nstat6_lastAck", "LAST_ACK v6"},
		{"nstat4_unknown", "UNKNOWN v4"},
		{"nstat6_unknown", "UNKNOWN v6"},
	}

	for i, state := range states {

		alias := fmt.Sprintf("pc%d", i)

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
				"LINE2:%s#%06X:%-14s",
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
			"%sconnections-passiveclose-%s.png",
			config.GlobalCfg.RRDHostnamePrefix,
			p.Name,
		),
	)

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         fmt.Sprintf("Passive Close Connections (%s)", p.Name),
		Start:         p.Start,
		VerticalLabel: "Connections",
		XGrid:         p.XGrid,
		Defs:          defs,
		Draw:          draw,
	}

	args := graph.BuildGraphArgs(t)

	if err := utils.ExecCommand(ctx, "CONNECTIONS", "rrdtool", args...); err != nil {
		logging.Error("CONNECTIONS", "Failed to create Passive Close graph '%s': %v", graphFile, err,)
		return
	}

	logging.Info("CONNECTIONS", "Created Passive Close graph '%s'", graphFile)
}
