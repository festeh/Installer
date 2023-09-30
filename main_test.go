package main

import (
	"testing"
)

var base = "/tmp"
var hostname = "host"

func TestGetConfigPath(t *testing.T) {
	_, err := GetConfigPath(base, hostname)
	if err != nil {
		t.Error(err)
	}
}

func TestLoadConfig(t *testing.T) {
	conf, _ := GetConfigPath(base, hostname)
	ParseConfig(conf)
}
