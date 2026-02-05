//
// internal/system/graph/meminfo.go
//
package graph

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"bufio"
	"fmt"
	
	"gonitorix/internal/config"
	"gonitorix/internal/graph"
	
)

func readMemTotal() (uint64, error) {
	file, err := os.Open("/proc/meminfo")

	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)

			if len(fields) >= 2 {
				var val uint64

				_, err := fmt.Sscanf(fields[1], "%d", &val)
				if err != nil {
					return 0, err
				}

				return val, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return 0, fmt.Errorf("MemTotal not found")
}

func createMeminfo(p *graph.GraphPeriod) {
	// Generates RRD graphs for Memory.

	rrdFile := config.GlobalCfg.RRDPath + "/system.rrd"
	graphFile := config.GlobalCfg.GraphPath + "/mem_" + p.Name + ".png"

	totalMem, _   := readMemTotal()
	totalMemBytes := uint64(totalMem * 1024)
	totalMemMB    := uint64(totalMem / 1024)
	
	t := graph.GraphTemplate{
		Graph:         graphFile,
		Title:         fmt.Sprintf("Memory allocation (%s) (%dMB)", p.Name, totalMemMB),
    	Start:         p.Start,
    	VerticalLabel: "bytes",
    	Width:         450,
    	Height:        150,
    	XGrid:         p.XGrid,

		Defs: []string{
			fmt.Sprintf("DEF:mtotl=%s:system_mtotl:AVERAGE", rrdFile),
           	fmt.Sprintf("DEF:mbuff=%s:system_mbuff:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:mcach=%s:system_mcach:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:mfree=%s:system_mfree:AVERAGE", rrdFile),
           	fmt.Sprintf("DEF:macti=%s:system_macti:AVERAGE", rrdFile),
			fmt.Sprintf("DEF:minac=%s:system_minac:AVERAGE", rrdFile),
		},

		CDefs: []string{
			"CDEF:m_mtotl=mtotl,1024,*",
			"CDEF:m_mbuff=mbuff,1024,*",
			"CDEF:m_mcach=mcach,1024,*",
			"CDEF:m_mused=m_mtotl,mfree,1024,*,-,m_mbuff,-,m_mcach,-",
			"CDEF:m_macti=macti,1024,*",
			"CDEF:m_minac=minac,1024,*",
			"CDEF:allvalues=mtotl,mbuff,mcach,mfree,macti,minac,+,+,+,+,+",
		},

		Draw: []string{
			"AREA:m_mused#EE4444:Used",
			"AREA:m_mcach#44EE44:Cached",
			"AREA:m_mbuff#CCCCCC:Buffers",
			"AREA:m_macti#E29136:Active",
			"AREA:m_minac#448844:Inactive",

			"LINE2:m_minac#008800",
			"LINE2:m_macti#E29136",
			"LINE2:m_mbuff#CCCCCC",
			"LINE2:m_mcach#00EE00",
			"LINE2:m_mused#EE0000",

			"COMMENT: \\n",
		},
	}

	_, errStat := os.Stat(graphFile)

	// Remove the PNG file if it exists.
	if !os.IsNotExist(errStat) {
		os.Remove(graphFile)
	}

	args := graph.BuildGraphArgs(t)

	args = append(args,
		fmt.Sprintf("--upper-limit=%d", totalMemBytes),
				    "--lower-limit=0",
	                "--rigid",
	                "--base=1024",
    )

	cmd := exec.Command("rrdtool", args...)
	err := cmd.Run()		

	if err != nil {
		log.Printf("Error creating image %s: %v\n", graphFile, err)
	} 
}