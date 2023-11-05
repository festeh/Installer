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

type Cmd struct {
	Cmd string
}

type Config struct {
	Symlinks map[string]Symlink
	// Cmds     map[string]Cmd
}

// Function to pretty print a config
func (c Config) String() string {
	return fmt.Sprintf("Symlinks: %s\n", c.Symlinks)
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
