package graph

import (
	"gonitorix/internal/config"
)

type period struct {
    name  string
    start string
	xGrid string
}

func Create(cfg *config.Config) {
	// Creates all configured network interface graphs.

	var (
			daily  = period{
				name:  "daily", 
				start: "-1day",
				xGrid: "HOUR:1:HOUR:6:HOUR:6:0:%R",
			}

			weekly = period{
				name:  "weekly", 
				start: "-1week",
			}

			monthly = period{
				name:  "monthly", 
				start: "-1month",
			}

			yearly = period{
				name:  "yearly", 
				start: "-1year",
			}
	)

	createBytes(cfg, daily)
	createBytes(cfg, weekly)
	createBytes(cfg, monthly)
	createBytes(cfg, yearly)

	createErrors(cfg, daily)
	createErrors(cfg, weekly)
	createErrors(cfg, monthly)
	createErrors(cfg, yearly)

	createPackets(cfg, daily)
	createPackets(cfg, weekly)
	createPackets(cfg, monthly)
	createPackets(cfg, yearly)
}