//
// internal/net/graph/bytes.go
//
package graph

import (
	"os"
	"os/exec"
	"fmt"
	"log"

	"gonitorix/internal/config"
	"gonitorix/internal/graph"
)

func createBytes(p *graph.GraphPeriod) {
	// Generates RRD graphs for byte transmission rates of the configured 
	// network interfaces.

	rrdPath := config.GlobalCfg.RRDPath
	graphPath := config.GlobalCfg.GraphPath

	for _, iface := range config.NetIfCfg.Interfaces {
		rrdFile := rrdPath + "/" + iface.Name + ".rrd"
		graphFile := graphPath + "/" + iface.Name + "_bytes_" + p.Name + ".png"

		t := graph.GraphTemplate{
			Graph:         graphFile,
			Title:         iface.Description + " (" + p.Name + ")",
			Start:         p.Start,
			VerticalLabel: "Bytes/s",
			XGrid:         p.XGrid,

			Defs: []string{
				fmt.Sprintf("DEF:in=%s:bytes_in:AVERAGE", rrdFile),
           		fmt.Sprintf("DEF:out=%s:bytes_out:AVERAGE", rrdFile),
			},

			CDefs: []string{
				"CDEF:allvalues=in,out,+",
				"CDEF:B_in=in",
				"CDEF:B_out=out",
				"CDEF:K_in=B_in,1024,/",
				"CDEF:K_out=B_out,1024,/",
				"COMMENT: \\n",
			},

			Draw: []string{
				"AREA:B_in#44EE44:KB/s Input",
				"GPRINT:K_in:LAST:     Current\\: %5.0lf",
				"GPRINT:K_in:AVERAGE: Average\\: %5.0lf",
				"GPRINT:K_in:MIN:    Min\\: %5.0lf",
				"GPRINT:K_in:MAX:    Max\\: %5.0lf\\n",

				"AREA:B_out#4444EE:KB/s Output",
				"GPRINT:K_out:LAST:    Current\\: %5.0lf",
				"GPRINT:K_out:AVERAGE: Average\\: %5.0lf",
				"GPRINT:K_out:MIN:    Min\\: %5.0lf",
				"GPRINT:K_out:MAX:    Max\\: %5.0lf\\n",

				"AREA:B_out#4444EE:",
				"AREA:B_in#44EE44:",
				"LINE1:B_out#0000EE",
				"LINE1:B_in#00EE00",
				"COMMENT: \\n",
				"COMMENT: \\n",
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