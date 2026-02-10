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
 
package utils

import (
	"log"
	"os"
	"os/exec"
	"fmt"
	"strings"
	
	"gonitorix/internal/logging"
)

// GetHostname returns the configured hostname for this instance.
func GetHostname() string {
	host, err := os.Hostname()

	if err != nil {
		log.Printf("Failed to get hostname: %v\n", err)
		return ""
	}

	return host
}

// GetDefaultGateways discovers the system's default IPv4 and IPv6 gateways
// by invoking the `ip route show default` command. It returns a map where 
// each key is the gateway IP address and the corresponding value is the 
// network interface used to reach it.
//
// Duplicate gateway addresses are detected and logged, and only the first
// occurrence is kept.
//
// An error is returned if no default gateways can be found.
func GetDefaultGateways() (map[string]string, error) {
	defaultGwCmd := [][]string{
		{"route", "show", "default"},
		{"-6", "route", "show", "default"},
	}

	gateways := make(map[string]string)

	for _, args := range defaultGwCmd {
		if logging.DebugEnabled() {
			logging.Debug("UTILS", "Discovering default gateways with: ip %s", strings.Join(args, " "),)
		}

		cmd := exec.Command("ip", args...)

		out, err := cmd.CombinedOutput()

		if err != nil {
			logging.Error("UTILS", "ip %v failed: %v", args, err,)

			if logging.DebugEnabled() {
				logging.Debug("UTILS", "ip output: %s", string(out),)
			}

			continue
		}

		if logging.DebugEnabled() {
			logging.Debug("UTILS", "ip output: %s", string(out),)
		}

		lines := strings.Split(string(out), "\n")

		for _, line := range lines {
			fields := strings.Fields(line)

			if len(fields) < 5 {
				continue
			}

			var gw, iface string

			for i := 0; i < len(fields); i++ {
				switch fields[i] {
					case "via":
						if i+1 < len(fields) {
							gw = fields[i+1]
						}

					case "dev":
						if i+1 < len(fields) {
							iface = fields[i+1]
						}
					}
			}

			if gw == "" || iface == "" {
				if logging.DebugEnabled() {
					logging.Debug("UTILS", "Ignoring line (missing gw/dev): %s", line,)
				}

				continue
			}

			// Duplicate detection
			if oldIface, exists := gateways[gw]; exists {
				logging.Warn("UTILS", "Duplicate default gateway detected: %s (interfaces %s and %s)", gw, oldIface, iface,)
				continue
			}

			gateways[gw] = iface

			if logging.DebugEnabled() {
				logging.Debug("UTILS", "Discovered default gateway %s via %s", gw, iface,)
			}
		}
	}

	if len(gateways) == 0 {
		return nil, fmt.Errorf("no default gateways found")
	}

	return gateways, nil
}

// GetIfaceFromIP determines which network interface would be used to 
// reach the given destination IP or hostname by querying the system 
// routing table. It returns the interface name, or an empty string 
// if it cannot be resolved.
func GetIfaceFromIP(ip string) (string, error) {
	args := []string{"route", "get", ip}

	if logging.DebugEnabled() {
		logging.Debug("UTILS", "Resolving interface for IP %s with: ip %s",	ip, strings.Join(args, " "),)
	}

	cmd := exec.Command("ip", args...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		logging.Error("UTILS", "ip route get %s failed: %v", ip, err,)

		if logging.DebugEnabled() {
			logging.Debug("UTILS", "ip output: %s", string(out),)
		}

		return "", err
	}

	if logging.DebugEnabled() {
		logging.Debug("UTILS", "ip route output for %s: %s", ip, string(out),)
	}

	fields := strings.Fields(string(out))

	for i := 0; i < len(fields); i++ {
		if fields[i] == "dev" && i+1 < len(fields) {

			if logging.DebugEnabled() {
				logging.Debug("UTILS", "Resolved interface for %s: %s", ip, fields[i+1],)
			}

			return fields[i+1], nil
		}
	}

	logging.Warn("UTILS", "Could not determine interface for destination %s", ip,)

	return "", fmt.Errorf("interface not found for %s", ip)
}