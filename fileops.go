package main

import (
	"fmt"
	"os"
)

func RemoveFile(path string) error {
	fmt.Printf("Removing %s\n", path)
	return os.Remove(path)
}

func CreateSymlink(source, target string) error {
	fmt.Printf("Creating symlink %s -> %s\n", source, target)
	return os.Symlink(source, target)
}
