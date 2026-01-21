package main

import (
	"os"
	"log"
	"flag"
	"os/exec"
	"context"

	"gonitorix/internal/config"
	"gonitorix/internal/net"
)

func startGonitorix(cfg *config.Config) {
	// Main Loop.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go net.Run(ctx, cfg)

	<-ctx.Done()

	os.Exit(0)
}

func main() {
	cfgFile := flag.String("c", "", "Configuration File")
	flag.Parse()

	if *cfgFile == "" {
		*cfgFile = "gonitorix.yaml"
	}

	// Verifies that the configuration file exists.
	_, errStat := os.Stat(*cfgFile)

	if os.IsNotExist(errStat) {
		log.Fatalf("The configuration file \"%s\" could not be opened or it does not exist.\n", *cfgFile)
		os.Exit(1)
	}

	// Verifies that the "rrdtool" binary is available.
	_, errLookPath := exec.LookPath("rrdtool")	

	if errLookPath != nil {
		log.Fatal("GONITORIX needs RRDtool installed to monitor your system.")
		os.Exit(1)
	}

	// Loads the configuration into a struct.
	cfg, err := config.Load(*cfgFile)

	if err != nil {
		log.Fatal(err)
	}

	startGonitorix(cfg)

	os.Exit(0)
}