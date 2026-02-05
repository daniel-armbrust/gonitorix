//
// internal/config/load.go
//
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
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		log.Fatalf("Cannot parse the configuration file %q: %w\n", cfgFile, err)
	}

	// Populate the application-wide configuration structures with the values
	// loaded from the configuration file.
	GlobalCfg = wrapper.Global
	SystemCfg = wrapper.System
	NetIfCfg = wrapper.NetIf

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