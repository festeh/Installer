package main

import (
	"github.com/BurntSushi/toml"
)


type Simple struct {
	Cmd   string `toml:"cmd"`
	Sudo  bool   `toml:"sudo" default:"false"`
	Check string `toml:"check"`
}

type InstallConfig struct {
	Simples map[string]Simple
}


func ParseInstallConfig(configPath string) (InstallConfig, error) {
	var config InstallConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return config, err
	}
	return config, nil
}
