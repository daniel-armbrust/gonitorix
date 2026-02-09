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
	"fmt"
	
	"gonitorix/internal/config"
	"gonitorix/internal/utils"
)



// prepareDefaultGateways discovers IPv4 and IPv6 default routes and adds them
// to the latency configuration, resolving the outgoing interface and RRD
// filename for each target.
func loadDefaultGateways() {
	var gateways map[string]string

	gateways, _ = utils.GetDefaultGateways()
		
	if len(gateways) == 0 {
		log.Println("Latency: no default gateways discovered")		
	} else {
		// Index existing addresses to avoid duplicates.
		existing := make(map[string]struct{})

		for _, host := range config.LatencyCfg.Hosts {
			existing[host.Address] = struct{}{}
		}
			
		for ip, iface := range gateways {
			// Skip duplicates already defined in YAML.
			if _, found := existing[ip]; found {
				log.Printf("Latency: gateway %s already configured, skipping\n", ip)
				continue
			}

			host := config.LatencyHost{
				Name:        "gateway-" + iface,
				Description: "Default gateway via " + iface,
				Address:     ip,
				Iface:       iface,
				RRDFile:     config.GlobalCfg.RRDHostnamePrefix + "latency_gw-" + iface + ".rrd",
			}

			config.LatencyCfg.Hosts = append(config.LatencyCfg.Hosts, host)

			log.Printf("Latency: discovered default gateway %s via %s and added to targets\n", ip, iface,)
		}
	}
}

// prepareTargets resolves routing and runtime details for latency monitoring.
//
// It populates each LatencyHost with the outgoing network interface and
// the RRD file path to be used for storage. The function also discovers
// default gateways when enabled and merges them into the host list.
//
// This must be called after configuration loading and before any RRD
// creation or ping execution.
func prepareLatencyTargets() {
	if config.LatencyCfg.DefaultGateway {
		loadDefaultGateways()
	}

	// Track known addresses to avoid duplicates.
	known := make(map[string]struct{})

	// Build a new filtered slice to safely remove invalid entries.
	filtered := config.LatencyCfg.Hosts[:0]

	for _, host := range config.LatencyCfg.Hosts {

		if host.Address == "" {
			continue
		}

		// Skip duplicate addresses.
		if _, exists := known[host.Address]; exists {
			continue
		}

		// If iface was not provided in YAML, try to discover
		// which local interface can reach this IP.
		if host.Iface == "" {
			iface, err := utils.GetIfaceFromIP(host.Address)

			if err != nil {
				log.Printf(
					"[WARN] Removing host %s (%s): %v",
					host.Name,
					host.Address,
					err,
				)
				continue
			}

			host.Iface = iface
		} 

		// Build RRD filename using prefix as part of the filename.
		host.RRDFile = fmt.Sprintf(
			"%slatency_%s.rrd",
			config.GlobalCfg.RRDHostnamePrefix,
			utils.SanitizeName(host.Name),
		)

		known[host.Address] = struct{}{}

		filtered = append(filtered, host)
	}

	config.LatencyCfg.Hosts = filtered
}