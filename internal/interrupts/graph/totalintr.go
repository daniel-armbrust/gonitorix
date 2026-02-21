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
	"path/filepath"

	"gonitorix/internal/config"
	"gonitorix/internal/utils"
	"gonitorix/internal/graph"
	"gonitorix/internal/logging"
)

func createTotalIntr(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "interrupts.rrd",
	)

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix + "interrupts-" + p.Name + ".png",
	)

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Total interrupts activity (" + p.Name + ")",
		Start:         p.Start,
		VerticalLabel: "Interrupts/s",
		XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:intr=%s:intr_total:AVERAGE", rrdFile),
		},

		Draw: []string{
			"AREA:intr#FFA500:Total",
			"LINE1:intr#FF8C00",
			"GPRINT:intr:LAST:  Cur\\:%9.2lf",
			"GPRINT:intr:AVERAGE:  Avg\\:%9.2lf",
			"GPRINT:intr:MAX:  Max\\:%9.2lf\\n",
		},
	}

	// Remove existing PNG if present
	if _, err := os.Stat(graphFile); err == nil {
		if err := os.Remove(graphFile); err != nil {
			logging.Warn("INTERRUPTS", "Failed to remove existing graph %s: %v", graphFile, err)
		}
	}

	args := graph.BuildGraphArgs(t)
	
	args = append(args,
		"--lower-limit=0",
		"--rigid",
	)

	if err := utils.ExecCommand(ctx, "INTERRUPTS", "rrdtool", args...); err != nil {
		logging.Error("INTERRUPTS", "Failed to create interrupts graph '%s': %v", graphFile, err)
		return
	}

	logging.Info("INTERRUPTS", "Created interrupts graph '%s'", graphFile)
}