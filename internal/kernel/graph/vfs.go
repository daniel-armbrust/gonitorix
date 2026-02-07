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

func createVfs(p *graph.GraphPeriod) {
	rrdFile := config.GlobalCfg.RRDPath + "/kernel.rrd"
	graphFile := config.GlobalCfg.GraphPath + "/kernvfs_" + p.Name + ".png"

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "VFS usage (" + p.Name + ")",
    	Start:         p.Start,
    	VerticalLabel: "Percent (%)",
    	XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:dentry=%s:kern_dentry:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:file=%s:kern_file:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:inode=%s:kern_inode:AVERAGE", rrdFile),
		},

		CDefs: []string{
			"CDEF:allvalues=dentry,file,inode,+,+",
		},

		Draw: []string{
			"AREA:inode#4444EE:inode",
			"GPRINT:inode:LAST:  Current\\: %4.1lf%%\\n",

			"AREA:dentry#EEEE44:dentry",
			"GPRINT:dentry:LAST: Current\\:  %4.1lf%%\\n",

			"AREA:file#EE44EE:file",
			"GPRINT:file:LAST:   Current\\:  %4.1lf%%\\n",

			"LINE2:inode#0000EE",
			"LINE2:dentry#EEEE00",
			"LINE2:file#EE00EE",
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