package main

import (
	"fmt"
	"testing"
)

func TestParseConfig(t *testing.T) {
	configPath := "test/test_config.toml"
	config, err := ParseConfig(configPath)
	if err != nil {
		t.Errorf("Error parsing config: %s", err)
	}
	if len(config.Templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(config.Templates))
	}
	fmt.Println(config)
}
