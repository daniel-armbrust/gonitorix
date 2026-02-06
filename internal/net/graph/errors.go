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
	"os/exec"
	"fmt"
	"log"

	"gonitorix/internal/config"
	"gonitorix/internal/graph"
)

func createErrors(p *graph.GraphPeriod) {
	// Creates error rate graphs for the configured network interfaces.
	rrdPath := config.GlobalCfg.RRDPath
	graphPath := config.GlobalCfg.GraphPath

	for _, iface := range config.NetIfCfg.Interfaces {
		rrdFile := rrdPath + "/" + iface.Name + ".rrd"
		graphFile := graphPath + "/" + iface.Name + "_errors_" + p.Name + ".png"

		t := graph.GraphTemplate{
			Graph:         graphFile,
			Title:         iface.Description + " (" + p.Name + ")",
			Start:         p.Start,
			VerticalLabel: "Errors/s",
			XGrid:         p.XGrid,

			Defs: []string{
				fmt.Sprintf("DEF:in=%s:errors_in:AVERAGE", rrdFile),
           		fmt.Sprintf("DEF:out=%s:errors_out:AVERAGE", rrdFile),
			},

			CDefs: []string{
				"CDEF:allvalues=in,out,+",
				"CDEF:e_in=in",
                "CDEF:e_out=out",
			},

			Draw: []string{
				"AREA:e_in#44EE44:Input",
                "AREA:e_out#4444EE:Output",
                "AREA:e_out#4444EE:",
                "AREA:e_in#44EE44:",
                "LINE1:e_out#0000EE", 
                "LINE1:e_in#00EE00",
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
}