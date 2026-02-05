//
// cmd/gonitorix/main.go
//
package main

import (
	"os"
	"os/exec"
	"log"	
	"context"
		
	"gonitorix/internal/config"
	"gonitorix/internal/system"
	"gonitorix/internal/net"
)

func startGonitorix() {
	// Main Loop.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// System Monitoring.
	go system.Run(ctx)

	// Network Monitoring
	go net.Run(ctx)

	<-ctx.Done()
}

func main() {
	cfgFile := "gonitorix.yaml"

	// Loads the configuration into a struct.
	config.Load(cfgFile)

	// Verifies that the "rrdtool" binary is available.
	_, errLookPath := exec.LookPath("rrdtool")	

	if errLookPath != nil {
		log.Fatalf("GONITORIX needs RRDtool installed to monitor your system.\n")
	}

	startGonitorix()

	os.Exit(0)
}