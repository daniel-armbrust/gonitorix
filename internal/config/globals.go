//
// internal/config/globals.go
//
package config

// --------------------
// GLOBAL
// --------------------

type GlobalConfig struct {
	RRDPath             string `yaml:"rrd_path"`
	GraphPath           string `yaml:"graph_path"`
	GraphHostnamePrefix bool   `yaml:"graph_hostname_prefix"`
}

type globalWrapper struct {
	Global GlobalConfig `yaml:"global"`
}

var GlobalCfg GlobalConfig

// --------------------
// SYSTEM
// --------------------

type SystemConfig struct {
	Step             int  `yaml:"step"`
	MaxHistoricYears int  `yaml:"max_historic_years"`
	CreateGraphs     bool `yaml:"create_graphs"`
}

type systemWrapper struct {
	System SystemConfig `yaml:"system"`
}

var SystemCfg SystemConfig

// --------------------
// NETWORK / NETIF
// --------------------

type NetIfConfig struct {
	Step             int            `yaml:"step"`
	MaxHistoricYears int            `yaml:"max_historic_years"`
	CreateGraphs     bool           `yaml:"create_graphs"`
	AutoDiscovery    bool			`yaml:"auto_discovery"`
	Interfaces       []NetInterface `yaml:"interfaces"`
}

type NetInterface struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Enable      bool   `yaml:"enable"`
}

type netIfWrapper struct {
	NetIf NetIfConfig `yaml:"netif"`
}

var NetIfCfg NetIfConfig