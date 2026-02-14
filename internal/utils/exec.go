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
	"bytes"
	"time"

	"gonitorix/internal/logging"
)

// ExecCommand executes an external command, logs it when debug mode is enabled,
// and returns the combined stdout/stderr output.
func ExecCommand(ctx context.Context, tag string, name string, args ...string,) error {
	cmd := exec.CommandContext(ctx, name, args...)

	if logging.DebugEnabled() {
		logging.Debug(tag, "Executing command: %s",	strings.Join(cmd.Args, " "),)
	}

	out, err := cmd.CombinedOutput()

	if err != nil && logging.DebugEnabled() {
		logging.Debug(tag, "Command output: %s", string(out),)
	}

	return err
}

// ExecCommandOutput executes a command with arguments and returns the
// combined stdout and stderr output as a string.
func ExecCommandOutput(ctx context.Context, tag string, cmd string, args ...string) (string, error) {
	if logging.DebugEnabled() {
		logging.Debug(tag, "Executing command: %s %v", cmd, args)
	}

	// Derive timeout from caller context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)

	var buf bytes.Buffer
	command.Stdout = &buf
	command.Stderr = &buf

	err := command.Run()

	out := buf.String()

	if ctx.Err() == context.DeadlineExceeded {
		if logging.DebugEnabled() {
			logging.Debug(tag, "Timeout running: %s %v", cmd, args)
		}
		return out, ctx.Err()
	}

	if err != nil && logging.DebugEnabled() {
		logging.Debug(tag, "Error: %v", err)
		logging.Debug(tag, "Output:\n%s", out)
	}

	if logging.DebugEnabled() && err == nil {
		logging.Debug(tag, "Output:\n%s", out)
	}

	return out, err
}