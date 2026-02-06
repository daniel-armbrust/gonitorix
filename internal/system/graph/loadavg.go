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

func createLoadavg(p *graph.GraphPeriod) {
	// Generates RRD graphs for Load Average.

	rrdFile := config.GlobalCfg.RRDPath + "/system.rrd"
	graphFile := config.GlobalCfg.GraphPath + "/loadavg_" + p.Name + ".png"
	
	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "System Load (" + p.Name + ")",
    	Start:         p.Start,
    	VerticalLabel: "Load average",
    	XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:load1=%s:system_load1:AVERAGE", rrdFile),
           	fmt.Sprintf("DEF:load5=%s:system_load5:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:load15=%s:system_load15:AVERAGE", rrdFile),
		},

		CDefs: []string{
			"CDEF:allvalues=load1,load5,load15,+,+",
		},

		Draw: []string{
			"AREA:load1#4444EE: 1 min average",

			"GPRINT:load1:LAST: Current\\: %4.2lf",
			"GPRINT:load1:AVERAGE: Average\\: %4.2lf",
			"GPRINT:load1:MIN: Min\\: %4.2lf",
			"GPRINT:load1:MAX: Max\\: %4.2lf\\n",

			"LINE1:load1#0000EE",
			"LINE1:load5#EEEE00: 5 min average",

			"GPRINT:load5:LAST: Current\\: %4.2lf",
			"GPRINT:load5:AVERAGE: Average\\: %4.2lf",
			"GPRINT:load5:MIN: Min\\: %4.2lf",
			"GPRINT:load5:MAX: Max\\: %4.2lf\\n",

			"LINE1:load15#00EEEE:15 min average",

			"GPRINT:load15:LAST: Current\\: %4.2lf",
			"GPRINT:load15:AVERAGE: Average\\: %4.2lf",
			"GPRINT:load15:MIN: Min\\: %4.2lf",
			"GPRINT:load15:MAX: Max\\: %4.2lf\\n",

			"COMMENT: \\n",
			"COMMENT:system uptime\\: 4 days, 17h 47m\\c",
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