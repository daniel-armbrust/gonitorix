//
// internal/net/stats.go
//
package net

import (
	"os"
	"bufio"
	"log"
	"fmt"
	"strings"
	"math"
	"strconv"

	"gonitorix/internal/config"
)

func discoveryIfaces() {
	file, err := os.Open("/proc/net/dev")

	if err != nil {
		log.Fatalf("Cannot read '/proc/net/dev': %w\n", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Skip first two header lines.
	for i := 0; i < 2 && scanner.Scan(); i++ {
	}

	found := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

	    parts := strings.SplitN(line, ":", 2)

		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])

		if iface == "" {
			continue
		}

		// Avoid duplicates.
		if found[iface] {
			continue
		}

		found[iface] = true

		config.NetIfCfg.Interfaces = append(
			config.NetIfCfg.Interfaces,
			config.NetInterface{
				Name:        iface,
				Description: fmt.Sprintf("Auto-discovered interface (%s)", iface),
				Enable:      true,
			},
		)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading '/proc/net/dev': %w\n", err)
	}
}

func parseFloat64(s string) float64 {
	// Parses a string into a float64 value, logging an error on failure, 
	// and rounds the result to 6 decimal places.

    v, err := strconv.ParseFloat(s, 64)

    if err != nil {
        log.Printf("Failed to parse float value %q\n", s)
        return 0
    }

    return math.Round(v*1e6) / 1e6
}

func roundFloat64(v float64) float64 {
	// Rounds a float64 value to a fixed precision of 6 decimal places.
    return math.Round(v*1e6) / 1e6
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

    return roundFloat64(delta / deltaT)
}

func computeRates(iface string, stats *ifStats, deltaT float64) ifStats {
	// Computes per-second transmission rates by comparing current interface 
	// counters with previously stored historical values.
	
	var rates ifStats

	for i := range config.NetIfCfg.Interfaces {
		if iface == config.NetIfCfg.Interfaces[i].Name {

			// Retrieves previously stored historical values.
			lastStats := lastIfstats[iface]

			rxBytes  := rate6(stats.rxBytes,  lastStats.rxBytes,  deltaT)
			txBytes  := rate6(stats.txBytes,  lastStats.txBytes,  deltaT)
			rxPkts   := rate6(stats.rxPkts,   lastStats.rxPkts,   deltaT)
			txPkts   := rate6(stats.txPkts,   lastStats.txPkts,   deltaT)
			rxErrors := rate6(stats.rxErrors, lastStats.rxErrors, deltaT)
			txErrors := rate6(stats.txErrors, lastStats.txErrors, deltaT)

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

func readStats() (map[string]*ifStats, error) {	
	// This function wil returns a map of structures holding 
	// per-interface network statistics (all network interface found).

	// Map that stores the data for all network interfaces 
	// read from /proc/net/dev
	procStats := make(map[string]*ifStats)

	file, err := os.Open("/proc/net/dev")

	if err != nil {
		log.Println("Cannot read '/proc/net/dev'")
		return nil, err
	}
	defer file.Close()

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
			log.Println("Invalid format in '/proc/net/dev' for '%s'", iface)
			continue
		}

		rxBytes  := parseFloat64(fields[0])
		rxPkts   := parseFloat64(fields[1])
		rxErrors := parseFloat64(fields[2])

		txBytes  := parseFloat64(fields[8])
		txPkts   := parseFloat64(fields[9])
		txErrors := parseFloat64(fields[10])
		
		// Structures that hold per-interface network statistics.
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
		log.Println("Cannot read '/proc/net/dev'")
	}

	return nil, err
}