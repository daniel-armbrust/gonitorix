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
	"gonitorix/internal/config"
	"gonitorix/internal/procfs"
	"gonitorix/internal/utils"
)

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

    return utils.RoundFloat64(delta / deltaT)
}

// computeRates calculates per-second rates from counter deltas between
// the current and previous samples.
func computeRates(iface string, stats *procfs.NetIfStat, deltaT float64) procfs.NetIfStat {
	// Computes per-second transmission rates by comparing current interface 
	// counters with previously stored historical values.
	
	var rates procfs.NetIfStat

	for i := range config.NetIfCfg.Interfaces {
		if iface == config.NetIfCfg.Interfaces[i].Name {

			// Retrieves previously stored historical values.
			lastStats := lastNetIfStats[iface]

			rxBytes  := rate6(stats.RxBytes,  lastStats.RxBytes,  deltaT)
			txBytes  := rate6(stats.TxBytes,  lastStats.TxBytes,  deltaT)
			rxPkts   := rate6(stats.RxPkts,   lastStats.RxPkts,   deltaT)
			txPkts   := rate6(stats.TxPkts,   lastStats.TxPkts,   deltaT)
			rxErrors := rate6(stats.RxErrors, lastStats.RxErrors, deltaT)
			txErrors := rate6(stats.TxErrors, lastStats.TxErrors, deltaT)

			rates = procfs.NetIfStat{
				RxBytes:  rxBytes,
				TxBytes:  txBytes,
				RxPkts:   rxPkts,
				TxPkts:   txPkts,						
				RxErrors: rxErrors,
				TxErrors: txErrors,
			}
		}
	}

	return rates
}