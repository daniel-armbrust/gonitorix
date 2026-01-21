package graph

import (
	"os"
	"os/exec"
	"fmt"
	"log"

	"gonitorix/internal/config"
)

func createErrors(cfg *config.Config, p period) {
	// Creates error rate graphs for the configured network interfaces.

	rrdPath := cfg.Global.RRDPath
	imgPath := cfg.Global.ImgPath

	for i, iface := range cfg.NetIf.Interfaces {
		rrdFile := rrdPath + "/" + iface.Config.Name + ".rrd"
		imgFile := imgPath + "/" + iface.Config.Name + "_errors_" + p.name + ".png"

		t := graphTemplate{
			img:           imgFile,
			title:         cfg.NetIf.Interfaces[i].Config.Description + " (" + p.name + ")",
    		start:         p.start,
    		verticalLabel: "Errors/s",
    		width:         450,
    		height:        150,
    		xGrid:         p.xGrid,

    		defs: []string{
				fmt.Sprintf("DEF:in=%s:errors_in:AVERAGE", rrdFile),
           		fmt.Sprintf("DEF:out=%s:errors_out:AVERAGE", rrdFile),
			},

			cdefs: []string{
				"CDEF:allvalues=in,out,+",
				"CDEF:e_in=in",
                "CDEF:e_out=out",
        	},

			draw: []string{
				"AREA:e_in#44EE44:Input",
                "AREA:e_out#4444EE:Output",
                "AREA:e_out#4444EE:",
                "AREA:e_in#44EE44:",
                "LINE1:e_out#0000EE", 
                "LINE1:e_in#00EE00",
			},
		}

		_, errStat := os.Stat(imgFile)

		// Remove the PNG file if it exists.
		if !os.IsNotExist(errStat) {
			os.Remove(imgFile)
		}

		args := buildGraphArgs(t)

		cmd := exec.Command("rrdtool", args...)
		err := cmd.Run()		

		if err != nil {
			log.Printf("Error creating image '%s' file: %v", imgFile, err)
		} 	
	}
}