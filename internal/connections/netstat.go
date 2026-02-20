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

package connections

import (
	"context"
	"strings"

	"gonitorix/internal/utils"
	"gonitorix/internal/logging"
)

 func collectFromNetstat(ctx context.Context) (connStats, connStats, error) {
	ipv4, err := collectNetstatFamily(ctx, "inet")

	if err != nil {
		return connStats{}, connStats{}, err
	}

	ipv6, err := collectNetstatFamily(ctx, "inet6")

	if err != nil {
		return connStats{}, connStats{}, err
	}

	return ipv4, ipv6, nil
}

func collectNetstatFamily(ctx context.Context, family string) (connStats, error) {
	var stats connStats

	// ---------------------------------------
	// TCP
	// ---------------------------------------
	args := []string{"-tn", "-A", family}

	output, err := utils.ExecCommandOutput(ctx, "CONNECTIONS", "netstat", args...)

	if err != nil {
		logging.Error("CONNECTIONS", "netstat -tn failed for %s", family)
		return connStats{}, err
	}

	parseNetstatTCP(output, &stats)

	// ---------------------------------------
	// LISTEN
	// ---------------------------------------
	args = []string{"-ltn", "-A", family}

	output, err = utils.ExecCommandOutput(ctx, "CONNECTIONS", "netstat", args...)

	if err != nil {
		logging.Error("CONNECTIONS", "netstat -ltn failed for %s", family)
		return connStats{}, err
	}

	parseNetstatListen(output, &stats)

	// ---------------------------------------
	// UDP
	// ---------------------------------------
	args = []string{"-lun", "-A", family}

	output, err = utils.ExecCommandOutput(ctx, "CONNECTIONS", "netstat", args...)

	if err != nil {
		logging.Error("CONNECTIONS", "netstat -lun failed for %s", family)
		return connStats{}, err
	}

	parseNetstatUDP(output, &stats, family)

	return stats, nil
}

func parseNetstatTCP(output string, stats *connStats) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) < 1 {
			continue
		}

		state := fields[len(fields)-1]

		switch state {
			case "CLOSED":
				stats.closed++
			case "SYN_SENT":
				stats.synSent++
			case "SYN_RECV":
				stats.synRecv++
			case "ESTABLISHED":
				stats.estab++
			case "FIN_WAIT1":
				stats.finWait1++
			case "FIN_WAIT2":
				stats.finWait2++
			case "CLOSING":
				stats.closing++
			case "TIME_WAIT":
				stats.timeWait++
			case "CLOSE_WAIT":
				stats.closeWait++
			case "LAST_ACK":
				stats.lastAck++
			case "UNKNOWN":
				stats.unknown++
		}
	}
}

func parseNetstatListen(output string, stats *connStats) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) < 1 {
			continue
		}

		state := fields[len(fields)-1]

		if state == "LISTEN" {
			stats.listen++
		}
	}
}

func parseNetstatUDP(output string, stats *connStats, family string) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) < 1 {
			continue
		}

		// IPv4: udp
		// IPv6: udp6
		if family == "inet" && strings.HasPrefix(line, "udp ") {
			stats.udp++
		}

		if family == "inet6" && strings.HasPrefix(line, "udp6") {
			stats.udp++
		}
	}
}
