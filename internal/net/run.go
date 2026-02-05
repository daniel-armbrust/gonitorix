//
// internal/net/run.go
//
package net

import (
	"context"
	"time"

	"gonitorix/internal/config"
)

func Run(ctx context.Context) {
	if config.NetIfCfg.AutoDiscovery {
		discoveryIfaces()
	}

	// Create RRD files.
	createRRD()

	// Call to updateNetIfStats routine to initialize the last values 
	// for calculating the differences. This way, the first update call 
	// will actually measure correct values.
	updateNetIfStats()
	
	ticker := time.NewTicker(time.Duration(config.NetIfCfg.Step) * time.Second)
	defer ticker.Stop()

	for {
		select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updateNetIfStats()
				
				// if cfg.NetIf.CreateGraphs {
				// 	graph.Create(cfg)
				// }
		}
	}
}