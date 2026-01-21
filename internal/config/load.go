package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

func Load(cfgFile string) (*Config, error) {	
	data, err := os.ReadFile(cfgFile)
	
	if err != nil {
		return nil, err
	}

	var raw struct {
		Global GlobalConfig  `yaml:"global"`
		NetIf  rawNetIfBlock `yaml:"netif"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	cfg := &Config{
		Global: raw.Global,
		NetIf: NetIfConfig{
			Step:             raw.NetIf.Step,
			MaxHistoricYears: raw.NetIf.MaxHistoricYears,
			CreateGraphs:     raw.NetIf.CreateGraphs,
			Interfaces:       BuildEnabledNetIfs(raw.NetIf),
		},
	}

	if err := cfg.NetIf.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
