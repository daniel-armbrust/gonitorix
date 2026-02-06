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
 
package config

import (
	"os"
	"log"

	"gopkg.in/yaml.v3"
)

func Load(cfgFile string) {
	data, err := os.ReadFile(cfgFile)

	if err != nil {
		log.Fatalf("The configuration file %q could not be opened: %w\n", cfgFile, err)
	}

	var wrapper struct {
		Global GlobalConfig `yaml:"global"`
		System SystemConfig `yaml:"system"`
		NetIf  NetIfConfig  `yaml:"netif"`
		Kernel KernelConfig `yaml:"kernel"`
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		log.Fatalf("Cannot parse the configuration file %q: %w\n", cfgFile, err)
	}

	// Populate the application-wide configuration structures with the values
	// loaded from the configuration file.
	GlobalCfg = wrapper.Global
	SystemCfg = wrapper.System
	NetIfCfg  = wrapper.NetIf
	KernelCfg = wrapper.Kernel

	// Network interface filtering logic.
	// If "auto_discovery" is disabled, keep only interfaces explicitly enabled 
	// in the configuration file.
	// If "auto_discovery" is enabled, keep all entries and let runtime 
	// discovery decide what is actually monitored.
	if NetIfCfg.AutoDiscovery {
		// Start empty — runtime discovery will populate it.
		NetIfCfg.Interfaces = nil
	} else {
		var enabled []NetInterface

		for _, iface := range NetIfCfg.Interfaces {
			if iface.Enable {
				enabled = append(enabled, iface)
			}
		}
		NetIfCfg.Interfaces = enabled
	}
}