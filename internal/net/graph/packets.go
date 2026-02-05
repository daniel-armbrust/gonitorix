//
// internal/net/graph/packets.go
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

func createPackets(p *graph.GraphPeriod) {
	// Creates packet traffic graphs for the configured network interfaces.

	rrdPath := config.GlobalCfg.RRDPath
	graphPath := config.GlobalCfg.GraphPath

	for iface := range cfg.NetIf.Interfaces {
		rrdFile := rrdPath + "/" + iface.Name + ".rrd"
		graphFile := graphPath + "/" + iface.Name + "_pkts_" + p.name + ".png"

		t := graphTemplate{
			Graph:         graphFile,
			Title:         iface.Description + " (" + p.name + ")",
			Start:         p.Start,
			VerticalLabel: "Packets/s",
			Width:         450,
			Height:        150,
			XGrid:         p.XGrid,

			Defs: []string{
				fmt.Sprintf("DEF:in=%s:packs_in:AVERAGE", rrdFile),
           		fmt.Sprintf("DEF:out=%s:packs_out:AVERAGE", rrdFile),
			},

			CDefs: []string{
				"CDEF:allvalues=in,out,+",
                "CDEF:p_in=in",
                "CDEF:p_out=out",
			},

			Draw: []string{
				"AREA:p_in#44EE44:Input",
                "AREA:p_out#4444EE:Output",
                "AREA:p_out#4444EE:",
                "AREA:p_in#44EE44:",
                "LINE1:p_out#0000EE", 
                "LINE1:p_in#00EE00",
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