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
	"context"
	"os/exec"
	"strings"

	"gonitorix/internal/logging"
)

// ExecCommand executes an external command, logs it when debug mode is enabled,
// and returns the combined stdout/stderr output.
func ExecCommand(ctx context.Context, tag string, name string, args ...string,) error {
	cmd := exec.CommandContext(ctx, name, args...)

	if logging.DebugEnabled() {
		logging.Debug(
			tag,
			"Executing command: %s",
			strings.Join(cmd.Args, " "),
		)
	}

	out, err := cmd.CombinedOutput()

	if err != nil && logging.DebugEnabled() {
		logging.Debug(
			tag,
			"Command output: %s",
			string(out),
		)
	}

	return err
}
