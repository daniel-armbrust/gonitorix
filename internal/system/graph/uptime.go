//
// internal/system/graph/uptime.go
//
package graph

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gonitorix/internal/config"
	"gonitorix/internal/graph"
)

type uptimeUnit struct {
	yTitle string
	unit   int
	format string
}

func uptimeUnitConfig(timeUnit string) uptimeUnit {
	switch strings.ToLower(timeUnit) {

	case "minute":
		return uptimeUnit{
			yTitle: "Minutes",
			unit:   60,
			format: "%5.0lf",
		}

	case "hour":
		return uptimeUnit{
			yTitle: "Hours",
			unit:   3600,
			format: "%5.0lf",
		}

	default:
		return uptimeUnit{
			yTitle: "Days",
			unit:   86400,
			format: "%5.1lf",
		}
	}
}

func createUptime(p *graph.GraphPeriod) {
	// Generates RRD graphs for machine Uptime.

	rrdFile := config.GlobalCfg.RRDPath + "/system.rrd"
	graphFile := config.GlobalCfg.GraphPath + "/uptime_" + p.Name + ".png"

	u := uptimeUnitConfig("")

	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         "Uptime (" + p.Name + ")",
    	Start:         p.Start,
    	VerticalLabel: u.yTitle,
    	Width:         450,
    	Height:        150,
    	XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:uptime=%s:system_uptime:AVERAGE", rrdFile),
		},

		CDefs: []string{
			fmt.Sprintf("CDEF:uptime_days=uptime,%d,/", u.unit),
			"CDEF:allvalues=uptime",
		},

		Draw: []string{
			"LINE2:uptime_days#EE44EE:Uptime",
			fmt.Sprintf(
				"GPRINT:uptime_days:LAST: Current\\:%s\\n",
				u.format,
			),
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