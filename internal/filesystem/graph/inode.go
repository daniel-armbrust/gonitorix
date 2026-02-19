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
	"strings"
	"path/filepath"
				
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/utils"
	"gonitorix/internal/graph"
)

func createInodeUsage(ctx context.Context, p *graph.GraphPeriod, devices []Device) {
	grouped := make(map[string][]Device)

	for _, d := range devices {
		grouped[d.RRDFile] = append(grouped[d.RRDFile], d)
	}

	for rrdFile, devs := range grouped {
		select {
			case <-ctx.Done():
				return
			default:
		}

		graphFile := filepath.Join(
			config.GlobalCfg.GraphPath,
			fmt.Sprintf(
				"%sfs-inode-%s.png",
				config.GlobalCfg.RRDHostnamePrefix,
				p.Name,
			),
		)

		t := graph.GraphTemplate{
			Graph:         graphFile,
			Title:         fmt.Sprintf("Inode usage (%s)", p.Name),
			Start:         p.Start,
			VerticalLabel: "Percent (%)",
			XGrid:         p.XGrid,
		}

		// ----------------------------
		// DEFs 
		// ----------------------------
		for _, d := range devs {
			t.Defs = append(t.Defs,
				fmt.Sprintf(
					"DEF:fs%d=%s:fs_ino%d:AVERAGE",
					d.Index,
					rrdFile,
					d.Index,
				),
			)
		}

		// ----------------------------
		// CDEF 
		// ----------------------------
		if len(devs) > 1 {
			cdef := "CDEF:allvalues="

			for i := 0; i < len(devs); i++ {
				cdef += fmt.Sprintf("fs%d,", i)
			}

			for i := 1; i < len(devs); i++ {
				cdef += "+,"
			}

			cdef = strings.TrimSuffix(cdef, ",")

			t.CDefs = append(t.CDefs, cdef)
		}

		// ----------------------------
		// DRAW
		// ----------------------------
		for _, d := range devs {
			colorInt := graph.GenerateHexColor(d.Index)
			color := fmt.Sprintf("#%06X", colorInt)

			t.Draw = append(t.Draw,
				fmt.Sprintf(
					"LINE2:fs%d%s:%s",
					d.Index,
					color,
					d.MountPoint,
				),
			)
		}

		args := graph.BuildGraphArgs(t)

		args = append(args,
			"--upper-limit=100",
			"--lower-limit=0",
			"--rigid",
		)

		if err := utils.ExecCommand(ctx, "FILESYSTEM", "rrdtool", args...); err != nil {
			logging.Error("FILESYSTEM", "Failed to create inode usage graph '%s': %v", graphFile, err,)
			continue
		}

		logging.Info("FILESYSTEM", "Created inode usage graph '%s'", graphFile,)
	}
}