/*
 * Gonitorix - a system and network monitoring tool
 * Copyright (C) 2026 Daniel Armbrust <darmbrust@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
 
package net

import (
	"os"
	"bufio"
	"fmt"
	"strings"
	"math"
	"strconv"
    "context"

	"gonitorix/internal/config"
	"gonitorix/internal/logging"
)

// discoveryIfaces scans /proc/net/dev and auto-discovers network interfaces,
// adding them to the runtime configuration when not explicitly defined.
// The operation can be cancelled through the provided context.
func discoveryIfaces(ctx context.Context) error {
	file, err := os.Open("/proc/net/dev")

	if err != nil {
		logging.Error("NET", "Cannot read /proc/net/dev: %v", err,)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Skip first two header lines.
	for i := 0; i < 2 && scanner.Scan(); i++ {
	}

	found := make(map[string]bool)

	for scanner.Scan() {
		select {
			case <-ctx.Done():
				logging.Info("NET", "Interface discovery cancelled")
				return ctx.Err()
			default:
		}

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

		if logging.DebugEnabled() {
			logging.Debug("NET", "Discovered interface %s",	iface,)
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("NET", "Error reading /proc/net/dev: %v", err,)
		return err
	}

	logging.Info("NET", "Discovered %d network interfaces", len(found),)

	return nil
}

// parseFloat64 converts a numeric string into float64.
// It returns 0 when the value cannot be parsed.
func parseFloat64(s string) float64 {
	// Parses a string into a float64 value, logging an error on failure, 
	// and rounds the result to 6 decimal places.

    v, err := strconv.ParseFloat(s, 64)

    if err != nil {
        logging.Error("NET", "Failed to parse float value %q\n", s)
        return 0
    }

    return math.Round(v*1e6) / 1e6
}

// roundFloat64 rounds a float64 value to the specified number of decimal 
// places.
func roundFloat64(v float64) float64 {
	// Rounds a float64 value to a fixed precision of 6 decimal places.
    return math.Round(v*1e6) / 1e6
}

// rate6 calculates a per-second rate and rounds the result to six decimal places.
// It is typically used for high-resolution network or system metrics.
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

// computeRates calculates per-second rates from counter deltas between
// the current and previous samples.
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

// readStats reads /proc/net/dev and returns a map containing per-interface
// network statistics such as bytes, packets and errors counters.
// The operation can be cancelled through the provided context.
func readStats(ctx context.Context) (map[string]*ifStats, error) {
	// Map that stores per-interface statistics read from /proc/net/dev.
	procStats := make(map[string]*ifStats)

	file, err := os.Open("/proc/net/dev")

	if err != nil {
		logging.Error("NET", "Cannot read /proc/net/dev: %v", err,)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
		}

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
			logging.Warn("NET",	"Invalid format in /proc/net/dev for interface %s",	iface,)
			continue
		}

		rxBytes := parseFloat64(fields[0])
		rxPkts := parseFloat64(fields[1])
		rxErrors := parseFloat64(fields[2])

		txBytes := parseFloat64(fields[8])
		txPkts := parseFloat64(fields[9])
		txErrors := parseFloat64(fields[10])

		// Store per-interface statistics.
		procStats[iface] = &ifStats{
			rxBytes:  rxBytes,
			txBytes:  txBytes,
			rxPkts:   rxPkts,
			txPkts:   txPkts,
			rxErrors: rxErrors,
			txErrors: txErrors,
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("NET", "Error reading /proc/net/dev: %v", err,)
		return nil, err
	}

	return procStats, nil
}