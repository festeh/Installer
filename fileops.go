package main

import (
	"log"
	"os"
	"path/filepath"
)

func RemoveFile(path string) error {
	log.Printf("Removing %s\n", path)
	return os.Remove(path)
}

func GetFiles(target string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(target, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		log.Printf("Error getting files: %s", err)
		return files, err
	}
	return files, nil
}

func isExistingDir(path string) bool {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return fi.IsDir()
}
