package config

import "errors"

// ---------- NETIF RUNTIME CONFIG ----------

type NetIfConfig struct {
	Step             int
	MaxHistoricYears int
	CreateGraphs     bool
	Interfaces       []NetIfInterface
}

// ---------- NETIF RAW (YAML) ----------

type rawNetIfBlock struct {
	Step             int                    `yaml:"step"`
	MaxHistoricYears int                    `yaml:"max_historic_years"`
	CreateGraphs	 bool					`yaml:"create_graphs"`
	Interfaces       []NetIfInterfaceConfig `yaml:"interfaces"`
}

type NetIfInterfaceConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Enable      bool   `yaml:"enable"`
}

// ---------- NETIF RUNTIME STRUCTS ----------

type NetIfStats struct {
	LastTimestamp float64
	RxBytes       float64
    TxBytes       float64
    RxPkts        float64
    TxPkts        float64
    RxErrors      float64
    TxErrors      float64
}

type NetIfInterface struct {
	Config NetIfInterfaceConfig
	Stats  *NetIfStats
}

func BuildEnabledNetIfs(raw rawNetIfBlock) []NetIfInterface {
	netifs := make([]NetIfInterface, 0, len(raw.Interfaces))

	for _, iface := range raw.Interfaces {
		if !iface.Enable {
			continue
		}

		netifs = append(netifs, NetIfInterface{
			Config: iface,
			Stats: &NetIfStats{
				LastTimestamp: 0,
				RxBytes:       0,
				TxBytes:       0,
				RxPkts:        0,
				TxPkts:        0,
				RxErrors:      0,
				TxErrors:      0,
			},
		})
	}

	return netifs
}

func (n *NetIfConfig) Validate() error {
	if n.Step == 0 {
		return errors.New("netif.step must be > 0.")
	}

	if n.MaxHistoricYears == 0 {
		return errors.New("netif.max_historic_years must be > 0.")
	}

	if len(n.Interfaces) == 0 {
		return errors.New("no enabled netif interfaces.")
	}
	
	return nil
}