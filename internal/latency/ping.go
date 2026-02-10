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
	"regexp"
	"fmt"
	"strconv"
	"time"
	"context"
	"strings"
	"os/exec"

	"gonitorix/internal/config"
	"gonitorix/internal/logging"
)

// parsePingOutput parses the output of the system ping command and extracts
// minimum, average, and maximum round-trip times as well as packet loss.
// It returns an error if the expected statistics cannot be found.
func parsePingOutput(out string) (*pingResult, error) {
	res := &pingResult{}

	// Packet loss
	lossRe := regexp.MustCompile(`(\d+(?:\.\d+)?)%\s+packet loss`)

	m := lossRe.FindStringSubmatch(out)

	if len(m) < 2 {
		return nil, fmt.Errorf("Could not parse packet loss\n")
	}

	res.loss, _ = strconv.ParseFloat(m[1], 64)

	// RTT line
	rttRe := regexp.MustCompile(`=\s*([\d\.]+)/([\d\.]+)/([\d\.]+)/`)

	m = rttRe.FindStringSubmatch(out)

	if len(m) < 4 {
		return nil, fmt.Errorf("Could not parse latency stats\n")
	}

	res.min, _ = strconv.ParseFloat(m[1], 64)
	res.avg, _ = strconv.ParseFloat(m[2], 64)
	res.max, _ = strconv.ParseFloat(m[3], 64)

	return res, nil
}

// pingProbe executes an ICMP ping to the given target using the specified
// timeout and packet count, optionally binding to a network interface.
// It parses the command output and returns latency statistics and packet
// loss percentages suitable for RRD updates.
func pingProbe(ctx context.Context, host config.LatencyHost, timeout time.Duration, packetCount int,) (*pingResult, error) {
	args := []string{}

	// Bind to interface if provided (Linux).
	if host.Iface != "" {
		args = append(args, "-I", host.Iface)
	}

	args = append(
		args,
		"-4",
		"-q",
		"-n",
		"-U",
		"-c", strconv.Itoa(packetCount),
		"-W", strconv.Itoa(int(timeout.Seconds())),
		host.Address,
	)

	if logging.DebugEnabled() {
		logging.Debug("LATENCY", "PING command: ping %s", strings.Join(args, " "),)
	}

	cmd := exec.CommandContext(ctx, "ping", args...)

	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		logging.Warn("LATENCY", "PING error for %s: %v", host.Address, err,)

		if logging.DebugEnabled() {
			logging.Debug("LATENCY", "PING output:\n%s", output,)
		}
	}

	return parsePingOutput(output)
}
