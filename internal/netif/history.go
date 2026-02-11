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
	"gonitorix/internal/procfs"
)

// filterNetIfStatsByConfig filters the collected network interface statistics
// and returns only those that are enabled in the configuration file.
func filterNetIfStatsByConfig(all map[string]*procfs.NetIfStats,) map[string]*procfs.NetIfStats {
	filtered := make(map[string]*procfs.NetIfStats)

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
func readNetIfStatsAndStoreHistory(ctx context.Context) {
	// Collect per-interface counters.
	procStats, err := procfs.ReadNetIfStats(ctx)

	if err != nil {
		logging.Warn("NETIF", "Failed to read interface statistics: %v", err,)
		return
	} 
	
	if !config.NetIfCfg.AutoDiscovery {
	    procStats = filterNetIfStatsByConfig(procStats) 
	} 

	// High resolution timestamp (seconds).
	timestamp := float64(time.Now().UnixNano()) / 1e9

	for iface, stats := range procStats {
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
			zeroStats := procfs.NetIfStats{
				RxBytes:  0,
				TxBytes:  0,
				RxPkts:   0,
				TxPkts:   0,
				RxErrors: 0,
				TxErrors: 0,
			}

			if err := updateRRD(ctx, rrdFile, &zeroStats); err != nil {
				logging.Warn("NETIF", "RRD initial update failed for %s: %v", iface, err,)
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
		lastNetIfstats[iface] = procfs.NetIfStats{
			RxBytes:  stats.RxBytes,
			TxBytes:  stats.TxBytes,
			RxPkts:   stats.RxPkts,
			TxPkts:   stats.TxPkts,
			RxErrors: stats.RxErrors,
			TxErrors: stats.TxErrors,
		}
	}

	// Save timestamp for next cycle.
	lastTimestamp = timestamp

	if logging.DebugEnabled() {
		logging.Debug("NETIF", "Network statistics updated for %d interfaces", len(procStats),)
	}
}