package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Simple struct {
	Cmd       string   `toml:"cmd"`
	Sudo      bool     `toml:"sudo" default:"false"`
	Check     string   `toml:"check"`
	DependsOn []string `toml:"depends_on"`
}

type InstallConfig struct {
	Inherits string `toml:"inherits"`
	Simples  map[string]Simple
}

func ParseInstallConfig(configPath string) (InstallConfig, error) {
	var config InstallConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return config, err
	}
	return config, nil
}

func LoadInstallConfig(basePath, hostname string) (InstallConfig, error) {
	return loadInstallConfigForHost(basePath, hostname, make(map[string]bool))
}

func loadInstallConfigForHost(basePath, host string, visited map[string]bool) (InstallConfig, error) {
	if visited[host] {
		return InstallConfig{}, fmt.Errorf("circular inheritance detected: %s", host)
	}
	visited[host] = true

	configPath := filepath.Join(basePath, "hosts", host, "install.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return InstallConfig{Simples: make(map[string]Simple)}, nil
	}

	config, err := ParseInstallConfig(configPath)
	if err != nil {
		return config, err
	}

	if config.Inherits == "" {
		return config, nil
	}

	parent, err := loadInstallConfigForHost(basePath, config.Inherits, visited)
	if err != nil {
		return config, fmt.Errorf("error loading parent '%s': %s", config.Inherits, err)
	}

	return mergeInstallConfigs(parent, config), nil
}

func mergeInstallConfigs(parent, child InstallConfig) InstallConfig {
	result := InstallConfig{Simples: make(map[string]Simple)}
	for k, v := range parent.Simples {
		result.Simples[k] = v
	}
	for k, v := range child.Simples {
		result.Simples[k] = v
	}
	return result
}
