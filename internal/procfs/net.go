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
 
package procfs

import (
	"os"
	"bufio"
	"fmt"
	"strings"
	"context"

	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/utils"
)

// ReadNetIfStats reads /proc/net/dev and returns a map containing per-interface
// network statistics such as bytes, packets and errors counters.
// The operation can be cancelled through the provided context.
func ReadNetIfStats(ctx context.Context) (map[string]*NetIfStats, error) {
	// Map that stores per-interface statistics read from /proc/net/dev.
	procNetIfStats := make(map[string]*NetIfStats)

	file, err := os.Open("/proc/net/dev")

	if err != nil {
		logging.Error("NETIF", "Cannot read /proc/net/dev: %v", err,)
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
			logging.Warn("NETIF", "Invalid format in /proc/net/dev for interface %s", iface,)
			continue
		}

		rxBytes := utils.ParseFloat64(fields[0])
		rxPkts := utils.ParseFloat64(fields[1])
		rxErrors := utils.ParseFloat64(fields[2])

		txBytes := utils.ParseFloat64(fields[8])
		txPkts := utils.ParseFloat64(fields[9])
		txErrors := utils.ParseFloat64(fields[10])

		// Store per-interface statistics.
		procNetIfStats[iface] = &NetIfStats{
			RxBytes:  rxBytes,
			TxBytes:  txBytes,
			RxPkts:   rxPkts,
			TxPkts:   txPkts,
			RxErrors: rxErrors,
			TxErrors: txErrors,
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("NETIF", "Error reading /proc/net/dev: %v", err,)
		return nil, err
	}

	return procNetIfStats, nil
}

// DiscoveryIfaces scans /proc/net/dev and auto-discovers network interfaces,
// adding them to the runtime configuration when not explicitly defined.
// The operation can be cancelled through the provided context.
func DiscoveryIfaces(ctx context.Context) error {
	file, err := os.Open("/proc/net/dev")

	if err != nil {
		logging.Error("NETIF", "Cannot read /proc/net/dev: %v", err,)
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
				logging.Info("NETIF", "Interface discovery cancelled")
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
			logging.Debug("NETIF", "Discovered interface %s",	iface,)
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Error("NETIF", "Error reading /proc/net/dev: %v", err,)
		return err
	}

	logging.Info("NETIF", "Discovered %d network interfaces", len(found),)

	return nil
}