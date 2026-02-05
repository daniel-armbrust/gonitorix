//
// internal/net/history.go
//
package net

import ( 
	"time"
	
	"gonitorix/internal/config"
)

func updateNetIfStats() {
	// Update logic for network interface RRD files, including the storage of
	// historical data required for subsequent updates.

	netIfStats, err := readStats()

	if err != nil {
		return
	}

	// Perl - Time::HiRes::time();
	timestamp := float64(time.Now().UnixNano()) / 1e9
	
	for iface, stats := range netIfStats {		
		rrdFile := config.GlobalCfg.RRDPath + "/" + iface + ".rrd"

		// lastTimestamp is a global package-level variable.
		if lastTimestamp == 0 {			
			zeroStats := ifStats{
				rxBytes:  0,
				txBytes:  0,
				rxPkts:   0,
				txPkts:   0,						
				rxErrors: 0,
				txErrors: 0,
			}

			// The first update is performed with zero values.
			updateRRD(rrdFile, &zeroStats)	
		} else {
			// Compute the elapsed time (deltaT) since the previous polling cycle.
			deltaT := timestamp - lastTimestamp

			// Compute rates and save it in history.
			rates := computeRates(iface, stats, deltaT)

			updateRRD(rrdFile, &rates)
		}

		// lastTimestamp stores the timestamp of the most recent polling 
		// cycle in a package-level variable for delta time calculations.
		lastTimestamp = timestamp

		// Store the current snapshot as the previous statistics for the next
		// polling cycle in a package-level map.
		lastIfstats[iface] = ifStats{
			rxBytes:  stats.rxBytes,
			txBytes:  stats.txBytes,	
			rxPkts:   stats.rxPkts,
			txPkts:   stats.txPkts,
			rxErrors: stats.rxErrors,
			txErrors: stats.txErrors,
		}		
	}	
}
