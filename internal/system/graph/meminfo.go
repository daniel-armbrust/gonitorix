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
	"gonitorix/internal/graph"
	"gonitorix/internal/utils"
	"gonitorix/internal/logging"	
	"gonitorix/internal/procfs"	
)

// createMeminfo generates RRD graphs showing system memory allocation
// for the given graph period.
func createMeminfo(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := filepath.Join(
		config.GlobalCfg.RRDPath,
		config.GlobalCfg.RRDHostnamePrefix + "system.rrd",
	)

	graphFile := filepath.Join(
		config.GlobalCfg.GraphPath,
		config.GlobalCfg.RRDHostnamePrefix + "mem-" + p.Name + ".png",
	)

	totalMemKB, err := procfs.ReadMemTotal(ctx)

	if err != nil {
		logging.Warn("SYSTEM", "Unable to read total memory: %v", err,)
		return
	}

	totalMemBytes := totalMemKB * 1024
	totalMemMB := totalMemKB / 1024

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         fmt.Sprintf("Memory Allocation (%s) (%dMB)", p.Name, totalMemMB),
		Start:         p.Start,
		VerticalLabel: "Bytes",
		XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:mtotl=%s:system_mtotl:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:mbuff=%s:system_mbuff:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:mcach=%s:system_mcach:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:mfree=%s:system_mfree:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:macti=%s:system_macti:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:minac=%s:system_minac:AVERAGE", rrdFile),
		},

		CDefs: []string{
			"CDEF:m_mtotl=mtotl,1024,*",
			"CDEF:m_mbuff=mbuff,1024,*",
			"CDEF:m_mcach=mcach,1024,*",
			"CDEF:m_mused=m_mtotl,mfree,1024,*,-,m_mbuff,-,m_mcach,-",
			"CDEF:m_macti=macti,1024,*",
			"CDEF:m_minac=minac,1024,*",
			"CDEF:allvalues=mtotl,mbuff,mcach,mfree,macti,minac,+,+,+,+,+",
		},

		Draw: []string{
			"AREA:m_mused#EE4444:Used",
			"COMMENT: \\n",

			"AREA:m_mcach#44EE44:Cached",
			"COMMENT: \\n",

			"AREA:m_mbuff#CCCCCC:Buffers",
			"COMMENT: \\n",

			"AREA:m_macti#E29136:Active",
			"COMMENT: \\n",

			"AREA:m_minac#448844:Inactive",

			"LINE2:m_minac#008800",
			"LINE2:m_macti#E29136",
			"LINE2:m_mbuff#CCCCCC",
			"LINE2:m_mcach#00EE00",
			"LINE2:m_mused#EE0000",

			"COMMENT: \\n",
		},
	}

	// Remove the PNG file if it already exists.
	if _, err := os.Stat(graphFile); err == nil {
		if err := os.Remove(graphFile); err != nil {
			logging.Warn("SYSTEM", "Failed to remove existing graph %s: %v", graphFile,	err,)
		}
	}

	args := graph.BuildGraphArgs(t)

	// Custom limits based on total memory.
	args = append(
		args,
		fmt.Sprintf("--upper-limit=%d", totalMemBytes),
		"--lower-limit=0",
		"--rigid",
		"--base=1024",
	)

	if err := utils.ExecCommand(ctx, "SYSTEM", "rrdtool", args...,); err != nil {
		logging.Error("SYSTEM", "Error creating image %s", graphFile,)
	}
}