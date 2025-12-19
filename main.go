package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	log.Println("Started")
	var commandVal, hostnameVal string
	var dryRun, force bool
	flag.StringVar(&commandVal, "command", "config", "Command: config, install, doctor, or uninstall")
	flag.StringVar(&commandVal, "c", "config", "Command (shorthand)")
	flag.StringVar(&hostnameVal, "host", "", "Hostname to run command on (auto-detected if not provided)")
	flag.StringVar(&hostnameVal, "h", "", "Hostname (shorthand)")
	flag.BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	flag.BoolVar(&dryRun, "n", false, "Dry run (shorthand)")
	flag.BoolVar(&force, "force", false, "Force regenerate all templates")
	flag.BoolVar(&force, "f", false, "Force (shorthand)")
	flag.Parse()
	if hostnameVal == "" {
		detectedHostname, err := os.Hostname()
		if err != nil {
			log.Fatal("Error: Could not detect hostname and none provided")
		}
		hostnameVal = detectedHostname
		log.Printf("Using detected hostname: %s", hostnameVal)
	}
	opts := Options{DryRun: dryRun, Force: force}
	installer, err := NewManager(hostnameVal, opts)
	if err != nil {
		log.Fatal(err)
	}
	err = installer.Dispatch(commandVal)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n✓ Done!\n")
}
