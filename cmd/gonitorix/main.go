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

package main

import (
	"os"
	"os/exec"
	"log"	
	"context"
		
	"gonitorix/internal/config"
	"gonitorix/internal/system"
	"gonitorix/internal/net"
	"gonitorix/internal/kernel"
	"gonitorix/internal/latency"
)

func startGonitorix() {
	// Main Loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// System Monitoring
	if config.SystemCfg.Enable {
		go system.Run(ctx)
	}	

	// Network Monitoring
	if config.NetIfCfg.Enable {
		go net.Run(ctx)
	}

	// Kernel Monitoring
	if config.KernelCfg.Enable {
		go kernel.Run(ctx)
	}

	// Latency Monitoring
	if config.LatencyCfg.Enable {
		go latency.Run(ctx)
	}

	<-ctx.Done()
}

func main() {
	cfgFile := "gonitorix.yaml"

	// Loads the configuration into a struct
	config.Load(cfgFile)

	// Verifies that the "rrdtool" binary is available
	_, errLookPath := exec.LookPath("rrdtool")	

	if errLookPath != nil {
		log.Fatalf("GONITORIX needs RRDtool installed to monitor your system.\n")
	}

	startGonitorix()

	os.Exit(0)
}