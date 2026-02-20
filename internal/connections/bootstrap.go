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
	"os/exec"
	"fmt"

	"gonitorix/internal/logging"
)

func initConnectionsMonitoring() (string, error) {
	if _, err := exec.LookPath("ss"); err == nil {
		logging.Info("CONNECTIONS", "Using 'ss' for connections collection")
		return "ss", nil
	}

	if _, err := exec.LookPath("netstat"); err == nil {
		logging.Info("CONNECTIONS", "Using 'netstat' for connections collection")
		return "netstat", nil
	}

	return "", fmt.Errorf("neither 'ss' nor 'netstat' found in PATH")
}
