package graph

import (
	"os"
	"os/exec"
	"fmt"
	"log"

	"gonitorix/internal/config"
)

func createBytes(cfg *config.Config, p period) {
	// Generates RRD graphs for byte transmission rates of the configured 
	// network interfaces.

	rrdPath := cfg.Global.RRDPath
	imgPath := cfg.Global.ImgPath

	for i, iface := range cfg.NetIf.Interfaces {
		rrdFile := rrdPath + "/" + iface.Config.Name + ".rrd"
		imgFile := imgPath + "/" + iface.Config.Name + "_bytes_" + p.name + ".png"

		t := graphTemplate{
			img:           imgFile,
			title:         cfg.NetIf.Interfaces[i].Config.Description + " (" + p.name + ")",
    		start:         p.start,
    		verticalLabel: "Bytes/s",
    		width:         450,
    		height:        150,
    		xGrid:         p.xGrid,

    		defs: []string{
				fmt.Sprintf("DEF:in=%s:bytes_in:AVERAGE", rrdFile),
           		fmt.Sprintf("DEF:out=%s:bytes_out:AVERAGE", rrdFile),
			},

			cdefs: []string{
				"CDEF:allvalues=in,out,+",
				"CDEF:B_in=in",
				"CDEF:B_out=out",
				"CDEF:K_in=B_in,1024,/",
				"CDEF:K_out=B_out,1024,/",
				"COMMENT: \\n",
        	},

			draw: []string{
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

		_, errStat := os.Stat(imgFile)

		// Remove the PNG file if it exists.
		if !os.IsNotExist(errStat) {
			os.Remove(imgFile)
		}

		args := buildGraphArgs(t)

		cmd := exec.Command("rrdtool", args...)
		err := cmd.Run()		

		if err != nil {
			log.Printf("Error creating image '%s': %v", imgFile, err)
		} 	

		// Debug.
		//log.Println(strings.Join(cmd.Args, " "))
	}
}