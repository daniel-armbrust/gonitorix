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

func createKernelUsage(p *graph.GraphPeriod) {
	rrdFile := config.GlobalCfg.RRDPath + "/kernel.rrd"
	graphFile := config.GlobalCfg.GraphPath + "/kernusage_" + p.Name + ".png"

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Kernel Usage (" + p.Name + ")",
    	Start:         p.Start,
    	VerticalLabel: "Percent (%)",
    	XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:user=%s:kern_user:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:nice=%s:kern_nice:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:sys=%s:kern_sys:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:iow=%s:kern_iow:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:irq=%s:kern_irq:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:sirq=%s:kern_sirq:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:steal=%s:kern_steal:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:guest=%s:kern_guest:AVERAGE", rrdFile),
		},

		CDefs: []string{
			"CDEF:allvalues=user,nice,sys,iow,irq,sirq,steal,guest,+,+,+,+,+,+,+",
		},

		Draw: []string{
			"AREA:user#4444EE:user",
			"GPRINT:user:LAST:     Current\\: %4.1lf%%",
			"GPRINT:user:AVERAGE: Average\\: %4.1lf%%",
			"GPRINT:user:MIN: Min\\: %4.1lf%%",
			"GPRINT:user:MAX: Max\\: %4.1lf%%\\n",

			"AREA:nice#EEEE44:nice",
			"GPRINT:nice:LAST:     Current\\: %4.1lf%%",
			"GPRINT:nice:AVERAGE: Average\\: %4.1lf%%",
			"GPRINT:nice:MIN: Min\\: %4.1lf%%",
			"GPRINT:nice:MAX: Max\\: %4.1lf%%\\n",

			"AREA:sys#44EEEE:system",
			"GPRINT:sys:LAST:   Current\\: %4.1lf%%",
			"GPRINT:sys:AVERAGE: Average\\: %4.1lf%%",
			"GPRINT:sys:MIN: Min\\: %4.1lf%%",
			"GPRINT:sys:MAX: Max\\: %4.1lf%%\\n",

			"AREA:iow#EE44EE:I/O wait",
			"GPRINT:iow:LAST: Current\\: %4.1lf%%",
			"GPRINT:iow:AVERAGE: Average\\: %4.1lf%%",
			"GPRINT:iow:MIN: Min\\: %4.1lf%%",
			"GPRINT:iow:MAX: Max\\: %4.1lf%%\\n",

			"AREA:irq#888888:IRQ",
			"GPRINT:irq:LAST:      Current\\: %4.1lf%%",
			"GPRINT:irq:AVERAGE: Average\\: %4.1lf%%",
			"GPRINT:irq:MIN: Min\\: %4.1lf%%",
			"GPRINT:irq:MAX: Max\\: %4.1lf%%\\n",

			"AREA:sirq#E29136:softIRQ",
			"GPRINT:sirq:LAST:  Current\\: %4.1lf%%",
			"GPRINT:sirq:AVERAGE: Average\\: %4.1lf%%",
			"GPRINT:sirq:MIN: Min\\: %4.1lf%%",
			"GPRINT:sirq:MAX: Max\\: %4.1lf%%\\n",

			"AREA:steal#44EE44:steal",
			"GPRINT:steal:LAST:    Current\\: %4.1lf%%",
			"GPRINT:steal:AVERAGE: Average\\: %4.1lf%%",
			"GPRINT:steal:MIN: Min\\: %4.1lf%%",
			"GPRINT:steal:MAX: Max\\: %4.1lf%%\\n",

			"AREA:guest#448844:guest",
			"GPRINT:guest:LAST:    Current\\: %4.1lf%%",
			"GPRINT:guest:AVERAGE: Average\\: %4.1lf%%",
			"GPRINT:guest:MIN: Min\\: %4.1lf%%",
			"GPRINT:guest:MAX: Max\\: %4.1lf%%\\n",

			"LINE1:guest#1F881F",
			"LINE1:steal#00EE00",
			"LINE1:sirq#D86612",
			"LINE1:irq#CCCCCC",
			"LINE1:iow#EE00EE",
			"LINE1:sys#00EEEE",
			"LINE1:nice#EEEE00",
			"LINE1:user#0000EE",
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
		"--upper-limit=100",
		"--lower-limit=0",
		"--rigid",
	)

	cmd := exec.Command("rrdtool", args...)
	err := cmd.Run()		

	if err != nil {
		log.Printf("Error creating image %s: %v\n", graphFile, err)
	}
}