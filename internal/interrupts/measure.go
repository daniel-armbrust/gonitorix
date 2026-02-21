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

package interrupts

import (
	"context"
	
	"gonitorix/internal/procfs"
	"gonitorix/internal/logging"
)

func measure(ctx context.Context) {
	stats, err := procfs.ReadInterruptStat(ctx)

	if err != nil {
		logging.Error("INTERRUPTS", "Failed to read interrupt stats: %v", err)
		return
	}

	if stats == nil {
		logging.Warn("INTERRUPTS", "InterruptStat returned nil")
		return
	}

	if err := updateRRD(ctx, stats); err != nil {
		logging.Error("INTERRUPTS", "Failed to update RRD: %v", err)
		return
	}
}