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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"log"	
	"context"
	"os/signal"
	"syscall"
		
	"gonitorix/internal/config"
	"gonitorix/internal/logging"
	"gonitorix/internal/system"
	"gonitorix/internal/netif"
	"gonitorix/internal/kernel"
	"gonitorix/internal/latency"
	"gonitorix/internal/process"
)

var GonitorixVersion = "dev"

func startGonitorix() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logging.Warn("MAIN", "Received signal %s, shutting down...", sig)
		cancel()
	}()

	if config.SystemCfg.Enable {
		logging.Info("SYSTEM", "Starting system monitoring subsystem")
	}

	if config.NetIfCfg.Enable {
		logging.Info("NETWORK", "Starting network monitoring subsystem")
	}

	if config.KernelCfg.Enable {
		logging.Info("KERNEL", "Starting kernel monitoring subsystem")
	}

	if config.LatencyCfg.Enable {
		logging.Info("LATENCY", "Starting latency monitoring subsystem")
	}

	if config.ProcessCfg.Enable {
		logging.Info("PROCESS", "Starting process monitoring subsystem")
	}

	if config.SystemCfg.Enable {
		go system.Run(ctx)
	}

	if config.NetIfCfg.Enable {
		go netif.Run(ctx)
	}

	if config.KernelCfg.Enable {
		go kernel.Run(ctx)
	}

	if config.LatencyCfg.Enable {
		go latency.Run(ctx)
	}

	if config.ProcessCfg.Enable {
		go process.Run(ctx)
	}

	// Block until cancellation
	<-ctx.Done()

	logging.Info("MAIN", "Shutdown complete")
}

func main() {
	flag.Parse()

	// Configure logging level
	logging.SetDebug(*debug)

	if *showVersion {
		fmt.Println("GONITORIX version", GonitorixVersion)
		os.Exit(0)
	}

	logging.Info("MAIN", "Starting Gonitorix (version=%s, pid=%d)", GonitorixVersion, os.Getpid())

	if *debug {
		logging.Debug("MAIN", "Debug mode enabled")
	}
	
	config.Load(*cfgFile)

	_, errLookPath := exec.LookPath("rrdtool")

	if errLookPath != nil {
		log.Fatalf("GONITORIX needs RRDtool installed to monitor your system.\n")
	}

	startGonitorix()

	os.Exit(0)
}