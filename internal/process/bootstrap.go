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
 
package process

import (
	"fmt"
	"path/filepath"

	"gonitorix/internal/config"
	"gonitorix/internal/utils"
)

// Initializes the structure that holds process stat data.
func initProcessMonitoring() {
	for _, process := range config.ProcessCfg.Processes {
		safeFilename := utils.SanitizeName(process.Name)

		rrdFile := filepath.Join(config.GlobalCfg.RRDPath,
			fmt.Sprintf("%sprocess-%s.rrd", config.GlobalCfg.RRDHostnamePrefix, safeFilename),
		)

		processHistory[process.Name] = &processStat{
			rrdFile:  rrdFile,
		}
	}
}