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
	"log"
	"sync"
	"time"

	"gonitorix/internal/config"
)

// measure runs network latency probes for all configured targets with
// controlled parallelism and stores the results in RRD files.
func measure() {
	maxParallel := config.LatencyCfg.MaxParallelProbes
	timeout := time.Duration(config.LatencyCfg.ProbeTimeoutSecs) * time.Second
	packetCount := config.LatencyCfg.ProbePackets

	// Semaphore channel to limit concurrency.
	sem := make(chan struct{}, maxParallel)

	var wg sync.WaitGroup

	for _, host := range config.LatencyCfg.Hosts {
		wg.Add(1)

		go func(h config.LatencyHost) {
			defer wg.Done()

			// Acquire semaphore slot.
			sem <- struct{}{}
			defer func() { <-sem }()

			pingResult, err := pingProbe(h, timeout, packetCount)

			if err != nil {
				log.Printf(
					"[ERROR] Probe failed for %s: %v",
					h.Address,
					err,
				)
				return
			}

			updateRRD(host.RRDFile, pingResult)
		}(host)
	}

	wg.Wait()
}
