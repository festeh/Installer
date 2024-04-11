package main

import (
	"flag"
	"log"
)

func main() {
	log.Println("Started")
	command := flag.String("command", "config", "Either install or config")
	hostname := flag.String("host", "", "Hostname to run command on")
	flag.Parse()
	if *hostname == "" {
		log.Fatal("Error: Hostname is required")
	}
	installer, err := NewManager(*hostname)
	if err != nil {
		log.Fatal(err)
	}
	err = installer.Dispatch(*command)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Done")
}
