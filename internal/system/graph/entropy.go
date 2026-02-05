//
// internal/system/graph/entropy.go
//
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
    	Width:         450,
    	Height:        150,
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