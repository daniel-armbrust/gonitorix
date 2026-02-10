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
	"os"
	"context"
		
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/utils"
	"gonitorix/internal/graph"
)

func createVfs(ctx context.Context, p *graph.GraphPeriod) {
	rrdFile := config.GlobalCfg.RRDPath + "/" + 
	           config.GlobalCfg.RRDHostnamePrefix + "kernel.rrd"
			   
	graphFile := config.GlobalCfg.GraphPath + "/" + 
	             config.GlobalCfg.RRDHostnamePrefix + 
				 "kernvfs-" + p.Name + ".png"

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

	// Remove the PNG file if it already exists.
	if _, err := os.Stat(graphFile); err == nil {

		if err := os.Remove(graphFile); err != nil {
			logging.Warn("KERNEL", "Failed to remove existing graph %s: %v", graphFile, err,)
		}
	}

	args := graph.BuildGraphArgs(t)

	// Additional custom arguments used to generate this graph.
	args = append(args,	"--upper-limit=100", "--lower-limit=0",	"--rigid",)

	// Execute rrdtool graph
	if err := utils.ExecCommand(ctx, "KERNEL", "rrdtool", args...,); err != nil {
		logging.Error("KERNEL", "Error creating image %s: %v", graphFile, err,)
	}
}