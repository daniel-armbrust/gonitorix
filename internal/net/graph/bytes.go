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
	"context"
	"fmt"

	"gonitorix/internal/config"
	"gonitorix/internal/utils"
	"gonitorix/internal/logging"
	"gonitorix/internal/graph"
)

// createBytes generates RRD graphs showing per-interface byte transmission
// rates for the given time period.
func createBytes(ctx context.Context, p *graph.GraphPeriod) {
	// Generates RRD graphs for byte transmission rates of the configured
	// network interfaces.
	for _, iface := range config.NetIfCfg.Interfaces {
		select {
			case <-ctx.Done():
				logging.Info("NET", "Byte graph generation cancelled")
				return
			default:
		}

		rrdFile := config.GlobalCfg.RRDPath + "/" +
				   config.GlobalCfg.RRDHostnamePrefix + iface.Name + ".rrd"

		graphFile := config.GlobalCfg.GraphPath + "/" +
					 config.GlobalCfg.RRDHostnamePrefix + iface.Name +
					 "_bytes-" + p.Name + ".png"

		t := graph.GraphTemplate{
			Graph:         graphFile,
			Title:         iface.Description + " (" + p.Name + ")",
			Start:         p.Start,
			VerticalLabel: "Bytes/s",
			XGrid:         p.XGrid,

			Defs: []string{
				fmt.Sprintf("DEF:in=%s:bytes_in:AVERAGE", rrdFile),
				fmt.Sprintf("DEF:out=%s:bytes_out:AVERAGE", rrdFile),
			},

			CDefs: []string{
				"CDEF:allvalues=in,out,+",
				"CDEF:B_in=in",
				"CDEF:B_out=out",
				"CDEF:K_in=B_in,1024,/",
				"CDEF:K_out=B_out,1024,/",
				"COMMENT: \\n",
			},

			Draw: []string{
				"AREA:B_in#44EE44:KB/s Input",
				"GPRINT:K_in:LAST:     Current\\: %5.0lf",
				"GPRINT:K_in:AVERAGE: Average\\: %5.0lf",
				"GPRINT:K_in:MIN:    Min\\: %5.0lf",
				"GPRINT:K_in:MAX:    Max\\: %5.0lf\\n",

				"AREA:B_out#4444EE:KB/s Output",
				"GPRINT:K_out:LAST:    Current\\: %5.0lf",
				"GPRINT:K_out:AVERAGE: Average\\: %5.0lf",
				"GPRINT:K_out:MIN:    Min\\: %5.0lf",
				"GPRINT:K_out:MAX:    Max\\: %5.0lf\\n",

				"AREA:B_out#4444EE:",
				"AREA:B_in#44EE44:",
				"LINE1:B_out#0000EE",
				"LINE1:B_in#00EE00",
				"COMMENT: \\n",
				"COMMENT: \\n",
			},
		}

		// Remove the PNG file if it already exists.
		if _, err := os.Stat(graphFile); err == nil {
			if err := os.Remove(graphFile); err != nil {
				logging.Warn("NET",	"Failed to remove existing graph %s: %v", graphFile, err,)
			}
		}

		args := graph.BuildGraphArgs(t)

		if err := utils.ExecCommand(ctx, "NET",	"rrdtool", args...,); err != nil {
			logging.Error("NET", "Error creating image %s",	graphFile,)
		}
	}
}