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
	"context"
	"time"

	"gonitorix/internal/config"
	"gonitorix/internal/net/graph"
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
				
				if config.NetIfCfg.CreateGraphs {
					graph.Create()
				}
		}
	}
}