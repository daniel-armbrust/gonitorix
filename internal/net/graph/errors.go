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
	"os"
	"fmt"
	"context"

	"gonitorix/internal/config"
	"gonitorix/internal/utils"
	"gonitorix/internal/logging"
	"gonitorix/internal/graph"
)

// createErrors generates RRD graphs showing per-interface packet error
// rates for the given graph period.
func createErrors(ctx context.Context, p *graph.GraphPeriod) {
	// Creates error rate graphs for the configured network interfaces.
	for _, iface := range config.NetIfCfg.Interfaces {
		select {
			case <-ctx.Done():
				logging.Info("NET", "Error graph generation cancelled")
				return
			default:
		}

		rrdFile := config.GlobalCfg.RRDPath + "/" +
				   config.GlobalCfg.RRDHostnamePrefix + iface.Name + ".rrd"

		graphFile := config.GlobalCfg.GraphPath + "/" +
					 config.GlobalCfg.RRDHostnamePrefix + iface.Name +
					 "_errors-" + p.Name + ".png"

		t := graph.GraphTemplate{
			Graph:         graphFile,
			Title:         iface.Description + " (" + p.Name + ")",
			Start:         p.Start,
			VerticalLabel: "Errors/s",
			XGrid:         p.XGrid,

			Defs: []string{
				fmt.Sprintf("DEF:in=%s:errors_in:AVERAGE", rrdFile),
				fmt.Sprintf("DEF:out=%s:errors_out:AVERAGE", rrdFile),
			},

			CDefs: []string{
				"CDEF:allvalues=in,out,+",
				"CDEF:e_in=in",
				"CDEF:e_out=out",
			},

			Draw: []string{
				"AREA:e_in#44EE44:Input",
				"AREA:e_out#4444EE:Output",
				"AREA:e_out#4444EE:",
				"AREA:e_in#44EE44:",
				"LINE1:e_out#0000EE",
				"LINE1:e_in#00EE00",
			},
		}

		// Remove the PNG file if it already exists.
		if _, err := os.Stat(graphFile); err == nil {
			if err := os.Remove(graphFile); err != nil {
				logging.Warn("NET", "Failed to remove existing graph %s: %v", graphFile, err,)
			}
		}

		args := graph.BuildGraphArgs(t)

		if err := utils.ExecCommand(ctx, "NET", "rrdtool", args...,); err != nil {
			logging.Error("NET", "Error creating image %s",	graphFile,)
		}
	}
}