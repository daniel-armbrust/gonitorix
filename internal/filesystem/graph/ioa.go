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

func createIOActivity(ctx context.Context, p *graph.GraphPeriod, devices []Device) {
	if len(devices) == 0 {
		return
	}

	var defs  []string
	var cdefs []string
	var draw  []string

	for i, dev := range devices {
		alias := fmt.Sprintf("io%d", i)
		aliasClean := fmt.Sprintf("%s_clean", alias)

		// -----------------------------------------
		// DEF
		// -----------------------------------------
		defs = append(defs,
			fmt.Sprintf(
				"DEF:%s=%s:fs_ioa%d:AVERAGE",
				alias,
				dev.RRDFile,
				dev.Index,
			),
		)

		// -----------------------------------------
		// Remove UNKNOWN
		// -----------------------------------------
		cdefs = append(cdefs,
			fmt.Sprintf(
				"CDEF:%s=%s,UN,0,%s,IF",
				aliasClean,
				alias,
				alias,
			),
		)

		// -----------------------------------------
		// LINE + GPRINT 
		// -----------------------------------------
		draw = append(draw,
			fmt.Sprintf(
				"LINE2:%s#%06X:%s",
				aliasClean,
				graph.GenerateHexColor(i),
				dev.MountPoint,
			),
		)

		draw = append(draw,
			fmt.Sprintf(
				"GPRINT:%s:LAST:  Cur\\: %%6.0lf",
				aliasClean,
			),
		)

		draw = append(draw,
			fmt.Sprintf(
				"GPRINT:%s:MIN:   Min\\: %%6.0lf",
				aliasClean,
			),
		)

		draw = append(draw,
			fmt.Sprintf(
				"GPRINT:%s:MAX:   Max\\: %%6.0lf\\l",
				aliasClean,
			),
		)
	}

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		fmt.Sprintf(
			"%sfs-io-%s.png",
			config.GlobalCfg.RRDHostnamePrefix,
			p.Name,
		),
	)

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         fmt.Sprintf("Disk I/O activity (%s)", p.Name),
		Start:         p.Start,
		VerticalLabel: "Reads+Writes/s",
		XGrid:         p.XGrid,
		Defs:          defs,
		CDefs:         cdefs,
		Draw:          draw,
	}

	args := graph.BuildGraphArgs(t)

	if err := utils.ExecCommand(ctx, "FILESYSTEM", "rrdtool", args...); err != nil {
		logging.Error("FILESYSTEM",	"Failed to create I/O activity graph '%s': %v", graphFile, err,)
		return
	}

	logging.Info("FILESYSTEM", "Created I/O activity graph '%s'", graphFile,)
}