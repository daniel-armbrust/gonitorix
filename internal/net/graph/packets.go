package graph

import (
	"os"
	"os/exec"
	"fmt"
	"log"

	"gonitorix/internal/config"
)

func createPackets(cfg *config.Config, p period) {
	// Creates packet traffic graphs for the configured network interfaces.

	rrdPath := cfg.Global.RRDPath
	imgPath := cfg.Global.ImgPath

	for i, iface := range cfg.NetIf.Interfaces {
		rrdFile := rrdPath + "/" + iface.Config.Name + ".rrd"
		imgFile := imgPath + "/" + iface.Config.Name + "_pkts_" + p.name + ".png"

		t := graphTemplate{
			img:           imgFile,
			title:         cfg.NetIf.Interfaces[i].Config.Description + " (" + p.name + ")",
    		start:         p.start,
    		verticalLabel: "Packets/s",
    		width:         450,
    		height:        150,
    		xGrid:         p.xGrid,

    		defs: []string{
				fmt.Sprintf("DEF:in=%s:packs_in:AVERAGE", rrdFile),
           		fmt.Sprintf("DEF:out=%s:packs_out:AVERAGE", rrdFile),
			},

			cdefs: []string{
				"CDEF:allvalues=in,out,+",
                "CDEF:p_in=in",
                "CDEF:p_out=out",
        	},

			draw: []string{
				"AREA:p_in#44EE44:Input",
                "AREA:p_out#4444EE:Output",
                "AREA:p_out#4444EE:",
                "AREA:p_in#44EE44:",
                "LINE1:p_out#0000EE", 
                "LINE1:p_in#00EE00",
			},
		}

		_, errStat := os.Stat(imgFile)

		// Remove the PNG file if it exists.
		if !os.IsNotExist(errStat) {
			os.Remove(imgFile)
		}

		// Builds and returns the argument list used to generate an RRD graph.
		args := buildGraphArgs(t)

		cmd := exec.Command("rrdtool", args...)
		err := cmd.Run()		

		if err != nil {
			log.Printf("Error creating image '%s' file: %v", imgFile, err)
		} 	
	}
}