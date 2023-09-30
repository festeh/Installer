package main

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Nvim Action `toml:"nvim"`
}

func (c Config) String() string {
	return fmt.Sprintf("Vim: {%s}", c.Nvim)
}

type Action struct {
	ActionType string `toml:"action"`
	Dest       string `toml:"dest,omitempty"`
}

type Result interface{}

func ParseConfig(path string) ([]Result, error) {
	var config Config

	// Load and unmarshal the TOML file
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	// Read file into res
	reader := toml.NewDecoder(file)
	reader.Decode(&config)
	// Print config
	println(config)

	return nil, nil
}
