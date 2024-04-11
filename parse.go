package main

import (
	"log"

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
	log.Printf("Parsing install config: %s\n", configPath)
	var config InstallConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatal(err)
		return config, err
	}
	return config, nil
}
