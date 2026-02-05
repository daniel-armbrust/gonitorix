//
// internal/system/rrd.go
// 
package system

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
	rrdFile := rrdPath + "/system.rrd"

	_, err := os.Stat(rrdFile)

	if os.IsNotExist(err) {
		args := []string{
				"create", rrdFile,
				"--step", strconv.Itoa(config.SystemCfg.Step),

				// DS
				"DS:system_load1:GAUGE:120:0:U",
  				"DS:system_load5:GAUGE:120:0:U",
  				"DS:system_load15:GAUGE:120:0:U",
  				"DS:system_nproc:GAUGE:120:0:U",
  				"DS:system_npslp:GAUGE:120:0:U",
  				"DS:system_nprun:GAUGE:120:0:U",
  				"DS:system_npwio:GAUGE:120:0:U",
				"DS:system_npzom:GAUGE:120:0:U",
				"DS:system_npstp:GAUGE:120:0:U",
				"DS:system_npswp:GAUGE:120:0:U",
				"DS:system_mtotl:GAUGE:120:0:U",
				"DS:system_mbuff:GAUGE:120:0:U",
				"DS:system_mcach:GAUGE:120:0:U",
				"DS:system_mfree:GAUGE:120:0:U",
				"DS:system_macti:GAUGE:120:0:U",
				"DS:system_minac:GAUGE:120:0:U",
				// "DS:system_val01:GAUGE:120:0:U",
				// "DS:system_val02:GAUGE:120:0:U",
				// "DS:system_val03:GAUGE:120:0:U",
				"DS:system_entrop:GAUGE:120:0:U",
				"DS:system_uptime:GAUGE:120:0:U",

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
		for n := 1; n <= config.SystemCfg.MaxHistoricYears; n++ {
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

		log.Printf("Creating RRD '%s'", rrdFile)	
	} else {
		log.Printf("RRD '%s' already exists", rrdFile)
	}	
}

func updateRRD() {
	rrdPath := config.GlobalCfg.RRDPath
	rrdFile := rrdPath + "/system.rrd"

	memory, err := readMemory()

	if err != nil {
		log.Printf("readMemory failed: %w\n", err)
	}

	loadavg, err := readLoadAvg()

	if err != nil {
		log.Printf("readLoadAvg failed: %w\n", err)
	}

	entropy, err := readEntropy()

	if err != nil {
		log.Printf("readEntropy failed: %w\n", err)
	}

	procinfo, err := readProcInfo()

	if err != nil {
		log.Printf("readProcInfo failed: %w\n", err)
	}

	uptime, err := readUptime()

	if err != nil {
		log.Printf("readUptime failed: %w\n", err)
	}

	rrdata := fmt.Sprintf(
		"N:%.2f:%.2f:%.2f:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%.0f",

		// load
		loadavg["load1"],
		loadavg["load5"],
		loadavg["load15"],

		// processos
		procinfo["total"],
		procinfo["sleep"],
		procinfo["run"],
		procinfo["wio"],
		procinfo["zombie"],
		procinfo["stop"],
		procinfo["swap"],

		// memória
		memory["MemTotal"],
		memory["Buffers"],
		memory["Cached"],
		memory["MemFree"],
		memory["Active"],
		memory["Inactive"],

		// outros
		entropy,
		uptime,
	)

	cmd := exec.Command(
		"rrdtool", "update", rrdFile, rrdata,
	)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("RRDTOOL update failed: %v | output: %s\n", err, out)
	}
}