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

// createEntropy generates an RRD graph showing kernel entropy values
// for the given graph period.
func createEntropy(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "system.rrd",
	)

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix + "entropy-" + p.Name + ".png",
	)

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Entropy (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: "Size",
		XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:entropy=%s:system_entrop:AVERAGE", rrdFile),
		},

		CDefs: []string{
			"CDEF:allvalues=entropy",
		},

		Draw: []string{
			"LINE2:entropy#EEEE00:Entropy",
			"GPRINT:entropy:LAST:  Current\\:%5.0lf\\n",
		},
	}

	// Remove the PNG file if it already exists.
	if _, err := os.Stat(graphFile); err == nil {
		if err := os.Remove(graphFile); err != nil {
			logging.Warn("SYSTEM", "Failed to remove existing graph %s: %v", graphFile, err,)
		}
	}

	args := graph.BuildGraphArgs(t)

	if err := utils.ExecCommand(ctx, "SYSTEM", "rrdtool", args...,); err != nil {
		logging.Error("SYSTEM", "Failed to create system entropy graph '%s': %v", graphFile, err,)
	}

	logging.Info("SYSTEM", "Created system entropy graph '%s'", graphFile,)
}