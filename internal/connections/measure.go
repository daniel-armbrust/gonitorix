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
	
	"gonitorix/internal/logging"
)

func measure(ctx context.Context, statsCmd string) {
	var (
		ipv4 connStats
		ipv6 connStats
		err  error
	)

	if statsCmd == "ss" {
		ipv4, ipv6, err = collectFromSS(ctx)
	} else {
		ipv4, ipv6, err = collectFromNetstat(ctx)
	}

	if err != nil {
		logging.Error("CONNECTIONS", "Failed collecting connections: %v", err)
		return
	}

	// --------------------------------------------------
	// Update RRD
	// --------------------------------------------------
	if err := updateRRD(ctx, ipv4, ipv6); err != nil {
		logging.Error("CONNECTIONS", "Failed updating RRD: %v", err)
		return
	}
}