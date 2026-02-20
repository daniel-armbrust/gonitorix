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

func collectFromSS(ctx context.Context) (connStats, connStats, error) {
	ipv4, err := runSS(ctx, "inet")

	if err != nil {
		return connStats{}, connStats{}, err
	}

	ipv6, err := runSS(ctx, "inet6")

	if err != nil {
		return connStats{}, connStats{}, err
	}

	return ipv4, ipv6, nil
}

func runSS(ctx context.Context, family string) (connStats, error) {
	args := []string{
		"-naut",
		"-f", family,
	}

	output, err := utils.ExecCommandOutput(ctx, "CONNECTIONS", "ss", args...)

	if err != nil {
		logging.Error("CONNECTIONS", "SS command failed for family '%s'", family)
		return connStats{}, err
	}

	return parseSSOutput(output), nil
}

func parseSSOutput(output string) connStats {
	var stats connStats

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		proto := fields[0]
		state := fields[1]

		switch proto {
			case "tcp":
				switch state {
					case "LISTEN":
						stats.listen++
					case "ESTAB":
						stats.estab++
					case "TIME-WAIT":
						stats.timeWait++
					case "CLOSE-WAIT":
						stats.closeWait++
					case "FIN-WAIT-1":
						stats.finWait1++
					case "FIN-WAIT-2":
						stats.finWait2++
					case "SYN-SENT":
						stats.synSent++
					case "SYN-RECV":
						stats.synRecv++
					case "CLOSING":
						stats.closing++
					case "LAST-ACK":
						stats.lastAck++
					case "UNCONN":
						stats.closed++
					case "UNKNOWN":
						stats.unknown++
				}
			case "udp":
				stats.udp++
			}
	}

	return stats
}