package main

import (
	"log"
	"os"
	"path"
)

func RemoveFile(path string) error {
	log.Printf("Removing %s\n", path)
	return os.Remove(path)
}

func CreateSymlink(source, target string) error {
	log.Printf("Creating symlink %s -> %s\n", source, target)
	if _, err := os.Stat(target); os.IsNotExist(err) {
		log.Printf("Error: Target %s does not exist\n", target)
		return err
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		log.Printf("Warning: Symlink %s already exists, removing\n", source)
		err := RemoveFile(source)
		if err != nil {
			return err
		}
	}
	// Create the symlink directory with parents if it doesn't exist
	err := os.MkdirAll(path.Dir(source), 0755)
	if err != nil {
		return err
	}
	err = os.Symlink(target, source)
	if err != nil {
		log.Printf("Error creating symlink: %s\n", err)
		return err
	}
	return nil
}
