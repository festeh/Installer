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
	flag.StringVar(&commandVal, "command", "config", "Either install or config")
	flag.StringVar(&commandVal, "c", "config", "Either install or config (shorthand)")
	flag.StringVar(&hostnameVal, "host", "", "Hostname to run command on (auto-detected if not provided)")
	flag.StringVar(&hostnameVal, "h", "", "Hostname to run command on (shorthand)")
	flag.Parse()
	if hostnameVal == "" {
		detectedHostname, err := os.Hostname()
		if err != nil {
			log.Fatal("Error: Could not detect hostname and none provided")
		}
		hostnameVal = detectedHostname
		log.Printf("Using detected hostname: %s", hostnameVal)
	}
	installer, err := NewManager(hostnameVal)
	if err != nil {
		log.Fatal(err)
	}
	err = installer.Dispatch(commandVal)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n✓ Done!\n")
}
