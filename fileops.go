package main

import (
	"log"
	"os"
)

func RemoveFile(path string) error {
	log.Printf("Removing %s\n", path)
	return os.Remove(path)
}
