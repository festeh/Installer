package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	log.Printf("Started")
	command := flag.String("command", "config", "Either install or config")
	hostname := flag.String("host", "", "Hostname to run command on")
	flag.Parse()
	if *hostname == "" {
		log.Printf("Error: Hostname is required")
		os.Exit(1)
	}
	installer := NewInstaller(*hostname)
	err := installer.Dispatch(*command)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Done")
}
