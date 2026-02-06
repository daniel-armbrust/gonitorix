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
	"log"
	"os"
	"os/exec"
	
	"gonitorix/internal/config"
	"gonitorix/internal/graph"
)

func createProcInfo(p *graph.GraphPeriod) {
	// Generates RRD graphs for Active Processes.

	rrdFile := config.GlobalCfg.RRDPath + "/system.rrd"
	graphFile := config.GlobalCfg.GraphPath + "/proc_" + p.Name + ".png"

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Active Processes (" + p.Name + ")",
    	Start:         p.Start,
    	VerticalLabel: "Processes",
    	XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:nproc=%s:system_nproc:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npslp=%s:system_npslp:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:nprun=%s:system_nprun:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npwio=%s:system_npwio:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npzom=%s:system_npzom:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npstp=%s:system_npstp:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:npswp=%s:system_npswp:AVERAGE", rrdFile),
		},

		CDefs: []string{
			"CDEF:allvalues=nproc,npslp,nprun,npwio,npzom,npstp,npswp,+,+,+,+,+,+",
		},

		Draw: []string{
			"AREA:npslp#448844:Sleeping",
			"GPRINT:npslp:LAST: Current\\:%5.0lf\\n",

			"LINE2:npwio#EE44EE:Wait I/O",
			"GPRINT:npwio:LAST: Current\\:%5.0lf\\n",

			"LINE2:npzom#00EEEE:Zombie",
			"GPRINT:npzom:LAST:   Current\\:%5.0lf\\n",

			"LINE2:npstp#EEEE00:Stopped",
			"GPRINT:npstp:LAST:  Current\\:%5.0lf\\n",

			"LINE2:npswp#0000EE:Paging",
			"GPRINT:npswp:LAST:   Current\\:%5.0lf\\n",

			"LINE2:nprun#EE0000:Running",
			"GPRINT:nprun:LAST:  Current\\:%5.0lf\\n",

			"COMMENT: \\n",

			"LINE2:nproc#888888:Total Processes",
			"GPRINT:nproc:LAST: Current\\:%5.0lf\\n",
		},
	}

	_, errStat := os.Stat(graphFile)

	// Remove the PNG file if it exists.
	if !os.IsNotExist(errStat) {
		os.Remove(graphFile)
	}

	args := graph.BuildGraphArgs(t)

	cmd := exec.Command("rrdtool", args...)
	err := cmd.Run()		

	if err != nil {
		log.Printf("Error creating image %s: %v\n", graphFile, err)
	} 
}