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

func createEntropy(p *graph.GraphPeriod) {
	// Generates RRD graphs for Entropy.

	rrdFile := config.GlobalCfg.RRDPath + "/system.rrd"
	graphFile := config.GlobalCfg.GraphPath + "/entropy_" + p.Name + ".png"

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
			"GPRINT:entropy:LAST: Current\\:%5.0lf\\n",
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