//
// internal/net/rrd.go
//
package net

import (
	"os"
	"os/exec"
	"strconv"
	"log"
	"fmt"

	"gonitorix/internal/config"
)

func createRRD() {
	rrdPath := config.GlobalCfg.RRDPath

	for _, iface := range config.NetIfCfg.Interfaces {
		rrdFile := rrdPath + "/" + iface.Name + ".rrd"

		_, err := os.Stat(rrdFile)

		if os.IsNotExist(err) {
			args := []string{
					"create", rrdFile,
					"--step", strconv.Itoa(config.NetIfCfg.Step),

					// DS
					"DS:bytes_in:GAUGE:120:0:U",
					"DS:bytes_out:GAUGE:120:0:U",
					"DS:packs_in:GAUGE:120:0:U",
					"DS:packs_out:GAUGE:120:0:U",
					"DS:errors_in:GAUGE:120:0:U",
					"DS:errors_out:GAUGE:120:0:U",

					// DAILY
					"RRA:AVERAGE:0.5:1:1440",
					"RRA:MIN:0.5:1:1440",
					"RRA:MAX:0.5:1:1440",
					"RRA:LAST:0.5:1:1440",

					// WEEKLY
					"RRA:AVERAGE:0.5:30:336",
					"RRA:MIN:0.5:30:336",
					"RRA:MAX:0.5:30:336",
					"RRA:LAST:0.5:30:336",

					// MONTHLY
					"RRA:AVERAGE:0.5:60:744",
					"RRA:MIN:0.5:60:744",
					"RRA:MAX:0.5:60:744",
					"RRA:LAST:0.5:60:744",
			}

			// YEARLY
			for n := 1; n <= config.NetIfCfg.MaxHistoricYears; n++ {
				rows := strconv.Itoa(365 * n)

				args = append(args,
					"RRA:AVERAGE:0.5:1440:" + rows,
					"RRA:MIN:0.5:1440:"     + rows,
					"RRA:MAX:0.5:1440:"     + rows,
					"RRA:LAST:0.5:1440:"    + rows,
				)
			}

			cmd := exec.Command("rrdtool", args...)			
			_, err := cmd.CombinedOutput()

			if err != nil {
				log.Printf("Error creating RRD '%s': %v\n", rrdFile, err)
				return
			}

			log.Printf("Creating RRD '%s'\n", rrdFile)			
		} else {
			log.Printf("RRD '%s' already exists", rrdFile)
		}		
	}
}

func updateRRD(rrdFile string, stats *ifStats) {
	cmd := exec.Command(
		"rrdtool", "update", rrdFile,
			fmt.Sprintf(
				"N:%.6f:%.6f:%.6f:%.6f:%.6f:%.6f",
				stats.rxBytes,
				stats.txBytes,
				stats.rxPkts,
				stats.txPkts,
				stats.rxErrors,
				stats.txErrors,
			),
	)

	err := cmd.Run()

	if err != nil {
	   log.Printf("RRDTOOL update failed: %v | output: %s\n", rrdFile, err)
	}
}