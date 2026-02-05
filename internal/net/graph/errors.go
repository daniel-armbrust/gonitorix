//
// internal/net/graph/errors.go
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

func createErrors(p *graph.GraphPeriod) {
	// Creates error rate graphs for the configured network interfaces.
	rrdPath := config.GlobalCfg.RRDPath
	graphPath := config.GlobalCfg.GraphPath

	for iface := range cfg.NetIf.Interfaces {
		rrdFile := rrdPath + "/" + iface.Name + ".rrd"
		graphFile := graphPath + "/" + iface.Name + "_errors_" + p.name + ".png"

		t := graphTemplate{
			Graph:         graphFile,
			Title:         iface.Description + " (" + p.name + ")",
			Start:         p.Start,
			VerticalLabel: "Errors/s",
			Width:         450,
			Height:        150,
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