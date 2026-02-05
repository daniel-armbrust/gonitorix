//
// internal/system/run.go
// 
package system

import (
	"context"
	"time"

	"gonitorix/internal/config"
	"gonitorix/internal/system/graph"
)

func Run(ctx context.Context) {
	createRRD()
	
	ticker := time.NewTicker(time.Duration(config.SystemCfg.Step) * time.Second)
	defer ticker.Stop()

	for {
		select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updateRRD()

				if config.SystemCfg.CreateGraphs {
					graph.Create()
				}
		}
	}
}