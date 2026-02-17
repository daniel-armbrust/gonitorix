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
	"gonitorix/internal/utils"
	"gonitorix/internal/graph"
)

func createPing(ctx context.Context, p *graph.GraphPeriod) {
	for _, host := range config.LatencyCfg.Hosts {
		select {
			case <-ctx.Done():
				logging.Info("LATENCY", "Ping graph generation cancelled")
				return
			default:
		}

		// rrdFile := config.GlobalCfg.RRDPath + "/" + host.RRDFile
		rrdFile := filepath.Join(
			config.GlobalCfg.RRDPath,
			host.RRDFile,
		)

		graphFile := fmt.Sprintf(
			"%s/latency_%s-%s.png",
			config.GlobalCfg.GraphPath,
			utils.SanitizeName(host.Name),
			p.Name,
		)

		t := graph.GraphTemplate{
			Graph:         graphFile,
			Title:         host.Description + " (" + p.Name + ")",
			Start:         p.Start,
			VerticalLabel: "Latency (ms)",
			XGrid:         p.XGrid,

			Defs: []string{
				fmt.Sprintf("DEF:rtt_min=%s:min:MIN", rrdFile),
				fmt.Sprintf("DEF:rtt_avg=%s:avg:AVERAGE", rrdFile),
				fmt.Sprintf("DEF:rtt_max=%s:max:MAX", rrdFile),
				fmt.Sprintf("DEF:rtt_loss=%s:loss:AVERAGE", rrdFile),

				"VDEF:vmin=rtt_min,MINIMUM",
				"VDEF:vavg=rtt_avg,AVERAGE",
				"VDEF:vmax=rtt_max,MAXIMUM",
				"VDEF:vloss=rtt_loss,AVERAGE",
			},

			Draw: []string{
				"LINE1:rtt_min#00FF99:Minimum",
				`GPRINT:vmin:%1.3lfms\l`,

				"LINE1:rtt_max#FF3333:Maximum",
				`GPRINT:vmax:%1.3lfms\l`,

				"LINE2:rtt_avg#00BFFF:Average",
				`GPRINT:vavg:%1.3lfms\l`,

				// Blank line before loss
				"COMMENT: \\l",

				"COMMENT:Lost packets",
				`GPRINT:vloss:%1.0lf%%\l`,
			},
		}

		// Remove the PNG file if it already exists.
		if _, err := os.Stat(graphFile); err == nil {
			if err := os.Remove(graphFile); err != nil {
				logging.Warn("LATENCY",	"Failed to remove existing graph %s: %v", graphFile, err,)
			}
		}

		args := graph.BuildGraphArgs(t)

		if err := utils.ExecCommand(ctx, "LATENCY", "rrdtool", args...,	); err != nil {
			logging.Error("LATENCY", "Error creating image %s", graphFile,)
		}
	}
}
