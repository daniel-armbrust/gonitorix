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

// --------------------
// GLOBAL
// --------------------

type GlobalConfig struct {
	RRDPath           string `yaml:"rrd_path"`
	GraphPath         string `yaml:"graph_path"`
	GraphWidth        int    `yaml:"graph_width"`
	GraphHeight       int    `yaml:"graph_height"`
	HostnamePrefix    bool   `yaml:"hostname_prefix"`
	RRDHostnamePrefix string
}

type globalWrapper struct {
	Global GlobalConfig `yaml:"global"`
}

// --------------------
// SYSTEM
// --------------------

type SystemConfig struct {
	Enable			 bool `yaml:"enable"`
	Step             int  `yaml:"step"`
	MaxHistoricYears int  `yaml:"max_historic_years"`
	CreateGraphs     bool `yaml:"create_graphs"`
}

type systemWrapper struct {
	System SystemConfig `yaml:"system"`
}

// --------------------
// KERNEL
// --------------------

type KernelConfig struct {
	Enable			 bool `yaml:"enable"`
	Step             int  `yaml:"step"`
	MaxHistoricYears int  `yaml:"max_historic_years"`
	CreateGraphs     bool `yaml:"create_graphs"`
}

type kernelWrapper struct {
	Kernel KernelConfig `yaml:"kernel"`
}

// --------------------
// FILE SYSTEM (FS)
// --------------------

type FilesystemConfig struct {
	Enable           bool     `yaml:"enable"`
	Step             int      `yaml:"step"`
	MaxHistoricYears int      `yaml:"max_historic_years"`
	CreateGraphs     bool     `yaml:"create_graphs"`
	MountPoints		 []string `yaml:"mountpoints"`
}

type filesystemWrapper struct {
	Process FilesystemConfig `yaml:"filesystem"`
}

// --------------------
// PROCESSES
// --------------------

type ProcessConfig struct {
	Enable           bool           `yaml:"enable"`
	Step             int            `yaml:"step"`
	MaxHistoricYears int            `yaml:"max_historic_years"`
	CreateGraphs     bool           `yaml:"create_graphs"`
	Processes        []ProcessEntry `yaml:"processes"`
}

type ProcessEntry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type processWrapper struct {
	Process ProcessConfig `yaml:"process"`
}

// --------------------
// NETWORK / NETIF
// --------------------

type NetIfConfig struct {
	Enable			 bool           `yaml:"enable"`
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

// --------------------
// NETWORK / LATENCY
// --------------------

type LatencyConfig struct {
	Enable            bool          `yaml:"enable"`
	Step              int           `yaml:"step"`
	MaxHistoricYears  int           `yaml:"max_historic_years"`
	CreateGraphs      bool          `yaml:"create_graphs"`
	DefaultGateway    bool          `yaml:"default_gateway"`
	MaxParallelProbes int           `yaml:"max_parallel_probes"`
	ProbeTimeoutSecs  int		    `yaml:"probe_timeout_seconds"`
	ProbePackets	  int		    `yaml:"probe_packets"`
	Hosts             []LatencyHost `yaml:"hosts"`
}

type LatencyHost struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Address     string `yaml:"address"`
	Iface       string `yaml:"iface"`
	RRDFile     string
}

type latencyWrapper struct {
	Latency LatencyConfig `yaml:"latency"`
}

// --------------------
// NETWORK / CONNECTIONS
// --------------------

type ConnectionsConfig struct {
	Enable            bool `yaml:"enable"`
	Step              int  `yaml:"step"`
	MaxHistoricYears  int  `yaml:"max_historic_years"`
	CreateGraphs      bool `yaml:"create_graphs"`
}

type connectionsWrapper struct {
	Connections ConnectionsConfig `yaml:"connections"`
}