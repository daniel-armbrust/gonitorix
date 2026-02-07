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

func createContextSwitches(p *graph.GraphPeriod) {
	rrdFile := config.GlobalCfg.RRDPath + "/" + 
	           config.GlobalCfg.RRDHostnamePrefix + "kernel.rrd"
			   
	graphFile := config.GlobalCfg.GraphPath + "/" + 
	             config.GlobalCfg.RRDHostnamePrefix + 
				 "kernctx-" + p.Name + ".png"

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Context Switches and Forks (" + p.Name + ")",
    	Start:         p.Start,
    	VerticalLabel: "CS & forks/s",
    	XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:cs=%s:kern_cs:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:forks=%s:kern_forks:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:vforks=%s:kern_vforks:AVERAGE", rrdFile),			
		},

		CDefs: []string{
			"CDEF:allvalues=cs,forks,vforks,+,+",
		},

		Draw: []string{
			"AREA:cs#44AAEE:Context switches",
			"GPRINT:cs:LAST: Current\\: %6.0lf\\n",

			"AREA:forks#4444EE:Forks",
			"GPRINT:forks:LAST:            Current\\: %6.0lf\\n",

			"LINE1:cs#00EEEE",
			"LINE1:forks#0000EE",
		},
	}
	_, errStat := os.Stat(graphFile)

	// Remove the PNG file if it exists.
	if !os.IsNotExist(errStat) {
		os.Remove(graphFile)
	}

	args := graph.BuildGraphArgs(t)

	// Additional custom arguments used to generate this graph.
	args = append(args,
		"--upper-limit=1000",
		"--lower-limit=0",
	)

	cmd := exec.Command("rrdtool", args...)
	err := cmd.Run()		

	if err != nil {
		log.Printf("Error creating image %s: %v\n", graphFile, err)
	}	
}