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

package latency

import (
	"sync"
	"time"
	"context"
	
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
)

// measure runs network latency probes for all configured targets with
// controlled parallelism and stores the results in RRD files.
func measure(ctx context.Context) {
	maxParallel := config.LatencyCfg.MaxParallelProbes
	timeout := time.Duration(config.LatencyCfg.ProbeTimeoutSecs) * time.Second
	packetCount := config.LatencyCfg.ProbePackets

	// Semaphore channel to limit concurrency.
	sem := make(chan struct{}, maxParallel)

	var wg sync.WaitGroup

	for _, host := range config.LatencyCfg.Hosts {
		select {
			case <-ctx.Done():
				logging.Info("LATENCY", "Measurement cancelled")
				return
			default:
		}

		wg.Add(1)

		go func(h config.LatencyHost) {
			defer wg.Done()

			// Acquire semaphore slot or abort if canceled.
			select {
				case sem <- struct{}{}:
				case <-ctx.Done():
					return
			}

			defer func() { <-sem }()

			pingResult, err := pingProbe(ctx, h, timeout, packetCount,)

			if err != nil {
				logging.Warn("LATENCY", "Probe failed for %s: %v", h.Address, err,)
				return
			}

			if err := updateRRD(ctx, h.RRDFile, pingResult); err != nil {
				logging.Warn("LATENCY",	"RRD update failed for %s: %v",	h.Address, err,)
			}

		}(host)
	}

	wg.Wait()
}