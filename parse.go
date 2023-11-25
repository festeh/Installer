package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type Symlink struct {
	Source string `toml:"source"`
	Target string `toml:"target"`
}

type Config struct {
	Symlinks map[string]Symlink
}

// Function to pretty print a config
func (c Config) String() string {
	return fmt.Sprintf("Symlinks: %v\n", c.Symlinks)
}

type Simple struct {
	Cmd string `toml:"cmd"`
	Sudo bool `toml:"sudo" default:"false"`
	Check string `toml:"check"`
}

type InstallConfig struct {
	Simples map[string]Simple
}

func ParseConfig(configPath string) (Config, error) {
	log.Printf("Parsing config: %s\n", configPath)
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatal(err)
		return config, err
	}
	return config, nil
}

func ParseInstallConfig(configPath string) (InstallConfig, error) {
	log.Printf("Parsing install config: %s\n", configPath)
	var config InstallConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatal(err)
		return config, err
	}
	return config, nil
}
