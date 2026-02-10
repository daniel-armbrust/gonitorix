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
 
package netif

import ( 
	"time"
	"context"
		
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
)

// filterNetIfStatsByConfig filters the collected network interface statistics
// and returns only those that are enabled in the configuration file.
func filterNetIfStatsByConfig(all map[string]*ifStats,) map[string]*ifStats {
	filtered := make(map[string]*ifStats)

	for _, iface := range config.NetIfCfg.Interfaces {
		if !iface.Enable {
			continue
		}

		if stats, ok := all[iface.Name]; ok {
			filtered[iface.Name] = stats
		}
	}

	return filtered
}

// updateNetIfStats collects network interface counters, computes per-second
// transmission rates and updates the corresponding RRD databases.
// The operation can be cancelled through the provided context.
func updateNetIfStats(ctx context.Context) {
	// Collect per-interface counters.
	netIfStats, err := readStats(ctx)

	if err != nil {
		logging.Warn("NETIF", "Failed to read interface statistics: %v", err,)
		return
	} 
	
	if !config.NetIfCfg.AutoDiscovery {
	    netIfStats = filterNetIfStatsByConfig(netIfStats) 
	} 

	// High resolution timestamp (seconds).
	timestamp := float64(time.Now().UnixNano()) / 1e9

	for iface, stats := range netIfStats {
		select {
			case <-ctx.Done():
				logging.Info("NETIF", "Network stats update cancelled")
				return
			default:
		}

		rrdFile := config.GlobalCfg.RRDPath + "/" +
				   config.GlobalCfg.RRDHostnamePrefix + iface + ".rrd"

		// First iteration: initialize RRD with zero values.
		if lastTimestamp == 0 {
			zeroStats := ifStats{
				rxBytes:  0,
				txBytes:  0,
				rxPkts:   0,
				txPkts:   0,
				rxErrors: 0,
				txErrors: 0,
			}

			if err := updateRRD(ctx, rrdFile, &zeroStats); err != nil {
				logging.Warn("NETIF", "RRD initial update failed for %s: %v",	iface, err,)
				continue
			}
		} else {
			// Compute elapsed time since previous cycle.
			deltaT := timestamp - lastTimestamp

			// Compute rates and save in history.
			rates := computeRates(iface, stats, deltaT)

			if err := updateRRD(ctx, rrdFile, &rates); err != nil {
				logging.Warn("NETIF", "RRD update failed for %s: %v",	iface, err,)
				continue
			}
		}

		// Store snapshot for next delta computation.
		lastIfstats[iface] = ifStats{
			rxBytes:  stats.rxBytes,
			txBytes:  stats.txBytes,
			rxPkts:   stats.rxPkts,
			txPkts:   stats.txPkts,
			rxErrors: stats.rxErrors,
			txErrors: stats.txErrors,
		}
	}

	// Save timestamp for next cycle.
	lastTimestamp = timestamp

	if logging.DebugEnabled() {
		logging.Debug("NETIF", "Network statistics updated for %d interfaces", len(netIfStats),)
	}
}