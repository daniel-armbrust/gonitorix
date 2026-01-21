package net

import (
	"os"
	"bufio"
	"strings"
	"fmt"
		
	"gonitorix/internal/config"
	"gonitorix/internal/utils"
)

type ifStats struct {
	rxBytes   float64
	txBytes   float64	
	rxPkts    float64
	txPkts    float64
	rxErrors  float64
	txErrors  float64
}

func rate6(current, previous, deltaT float64) float64 {	
	// Calculates the transmission rate using the difference between 
	// current and previous counters, normalized over the given time 
	// interval and rounded to 6 decimal places.

	if deltaT <= 0 {
        return 0
    }

    if previous <= 0 {
        return 0
    }

    delta := current - previous
    if delta <= 0 {
        return 0
    }

    return utils.RoundFloat64(delta / deltaT)
}

func readStats() map[string]*ifStats {	
	file, err := os.Open("/proc/net/dev")

	if err != nil {
		panic("Cannot read \"/proc/net/dev\".")
	}
	defer file.Close()

	// Mapa que armazena os dados de todas as interfaces de rede 
	// lidos de /proc/net/dev.
	procStats := make(map[string]*ifStats)

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Skip headers.
		if lineNum <= 2 {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])

		// We need at least 16 fields.
		if len(fields) < 16 {
			fmt.Errorf("Invalid format in \"/proc/net/dev\" for \"%s\".", iface)
			continue
		}

		rxBytes  := utils.ParseFloat64(fields[0])
		rxPkts   := utils.ParseFloat64(fields[1])
		rxErrors := utils.ParseFloat64(fields[2])

		txBytes  := utils.ParseFloat64(fields[8])
		txPkts   := utils.ParseFloat64(fields[9])
		txErrors := utils.ParseFloat64(fields[10])

		procStats[iface] = &ifStats{
			rxBytes:   rxBytes,
			txBytes:   txBytes,
			rxPkts:    rxPkts,
			txPkts:    txPkts,
			rxErrors:  rxErrors,
			txErrors:  txErrors,
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Errorf("Cannot read \"/proc/net/dev\".")
	}

	return procStats
}

func computeRates(cfg *config.Config, 
				  iface string, 
				  stats *ifStats, 
				  deltaT float64) ifStats {
	// Computes per-second transmission rates by comparing current interface 
	// counters with previously stored historical values.
	
	var rates ifStats

	for i := range cfg.NetIf.Interfaces {
		if iface == cfg.NetIf.Interfaces[i].Config.Name {

			// Retrieves previously stored historical values.
			statsHist := cfg.NetIf.Interfaces[i].Stats

			rxBytes  := rate6(stats.rxBytes,  statsHist.RxBytes,  deltaT)
			txBytes  := rate6(stats.txBytes,  statsHist.TxBytes,  deltaT)
			rxPkts   := rate6(stats.rxPkts,   statsHist.RxPkts,   deltaT)
			txPkts   := rate6(stats.txPkts,   statsHist.TxPkts,   deltaT)
			rxErrors := rate6(stats.rxErrors, statsHist.RxErrors, deltaT)
			txErrors := rate6(stats.txErrors, statsHist.TxErrors, deltaT)

			rates = ifStats{
				rxBytes:  rxBytes,
				txBytes:  txBytes,
				rxPkts:   rxPkts,
				txPkts:   txPkts,						
				rxErrors: rxErrors,
				txErrors: txErrors,
			}

		}
	}

	return rates
}