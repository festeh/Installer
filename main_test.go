package main

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	configPath := "test/test_config.toml"
	config, err := ParseConfig(configPath)
	if err != nil {
		t.Errorf("Error parsing config: %s", err)
	}
	// Preppty print the config
	t.Log(config)
	if len(config.Symlinks) != 2 {
		t.Errorf("Expected 2 symlinks, got %d", len(config.Symlinks))
	}
}
