package net

import ( 
	"time"
	
	"gonitorix/internal/config"
)

func updateHistoryStats(cfg *config.Config, iface string, stats *ifStats) {
	// Updates the historical data for each network interface.
	// Maintaining historical data is required to convert raw counters read from
	// /proc/net/dev into transmission rates.

	for i := range cfg.NetIf.Interfaces {
		if iface == cfg.NetIf.Interfaces[i].Config.Name {			
			cfg.NetIf.Interfaces[i].Stats.RxBytes = stats.rxBytes
			cfg.NetIf.Interfaces[i].Stats.TxBytes = stats.txBytes

			cfg.NetIf.Interfaces[i].Stats.RxPkts = stats.rxPkts
			cfg.NetIf.Interfaces[i].Stats.TxPkts = stats.txPkts

			cfg.NetIf.Interfaces[i].Stats.RxErrors = stats.rxErrors
			cfg.NetIf.Interfaces[i].Stats.TxErrors = stats.txErrors
		}
	}	
}

func updateLogic(cfg *config.Config) {
	// Update logic for network interface RRD files, including the storage of
	// historical data required for subsequent updates.

	netIfStats := readStats()

	// Perl - Time::HiRes::time();
	timestamp := float64(time.Now().UnixNano()) / 1e9
	
	for iface, stats := range netIfStats {		
		for i, configIface := range cfg.NetIf.Interfaces {

			// Processes only the network interfaces defined in 
			// the "gonitorix.yaml" configuration file.
			if iface == configIface.Config.Name {
				rrdFile := cfg.Global.RRDPath + "/" + iface + ".rrd"

				if cfg.NetIf.Interfaces[i].Stats.LastTimestamp == 0 {

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

					lastTimestamp := cfg.NetIf.Interfaces[i].Stats.LastTimestamp
					deltaT := timestamp - lastTimestamp

					rates := computeRates(cfg, iface, stats, deltaT)							

					// Writes the calculated rate values to the RRD.
					updateRRD(rrdFile, &rates)
				}	

				// Stores data in the historical records.
				cfg.NetIf.Interfaces[i].Stats.LastTimestamp = timestamp	
				updateHistoryStats(cfg, iface, stats)				
			}
		}
	}	
}
