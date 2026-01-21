package net

import (
	"context"
	"time"

	"gonitorix/internal/config"
	"gonitorix/internal/net/graph"
)

func Run(ctx context.Context, cfg *config.Config) {
	createRRD(cfg)

	// Call to updateLogic routine to initialize the last values for calculating 
	// the differences. This way, the first update call will actually measure 
	// correct values.
	updateLogic(cfg)

	ticker := time.NewTicker(time.Duration(cfg.NetIf.Step) * time.Second)
	defer ticker.Stop()

	for {
		select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updateLogic(cfg)
				
				if cfg.NetIf.CreateGraphs {
					graph.Create(cfg)
				}
		}
	}
}