package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
)

// Builds a config and checks if it exists
func GetConfigPath(base string, hostname string) (string, error) {
	configPath := path.Join(base, hostname, "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Error: Config %s does not exist\n", configPath)
		return "", err
	}
	return configPath, nil
}

func ExpandPath(path string) (string, error) {
		usr, err := user.Current()
		if err != nil {
				return "", err
		}
		if path[:2] == "~/" {
				path = strings.Replace(path, "~", usr.HomeDir, 1)
		}
		return path, nil
}

func mainFunc(base string, hostname string) error {
	configPath, err := GetConfigPath(base, hostname)
	if err != nil {
		log.Printf("Error getting config path: %s", err)
		return err
	}
	config, err := ParseConfig(configPath)
	for _, symlink := range config.Symlinks {
		target := path.Join(base, symlink.Target)
		if _, err := os.Stat(target); os.IsNotExist(err) {
			log.Printf("Error: Target %s does not exist\n", target)
			return err
		}
		source, err := ExpandPath(symlink.Source)
		if err != nil {
			log.Printf("Error expanding path: %s", err)
			return err
		}
		err = CreateSymlink(source, target)
		if err != nil {
			return err
		}
		fmt.Println("")
	}
	return nil
}

func main() {
	log.Printf("Starting dotfiles install")
	hostname := os.Args[1]
	log.Printf("Installing for hostname: %s", hostname)
	base := os.Getenv("HOME") + "/dotfiles"
	err := mainFunc(base, hostname)
	if err != nil {
		log.Fatal(err)
		// exit with non-zero status
		os.Exit(1)
	}
	log.Printf("Finished dotfiles install")
}
