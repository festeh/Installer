package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Manager struct {
	base     string
	hostname string
}

func NewManager(hostname string) (*Manager, error) {
	if err := IsHostnameOk(hostname); err != nil {
		return nil, err
	}
	base := os.Getenv("HOME") + "/dotfiles"
	expanded, err := ExpandHomeDir(base)
	if err != nil {
		return nil, err
	}
	return &Manager{hostname: hostname, base: expanded}, nil
}

func (i *Manager) Dispatch(command string) error {
	if command == "install" {
		// TODO: Implement install
		return nil
	} else if command == "config" {
		configurer := NewConfigurer(i.hostname, i.base)
		return configurer.Run()
	} else {
		return fmt.Errorf("Unknown command: %s", command)
	}
}

func IsHostnameOk(hostname string) error {
	if hostname == "common" {
		return nil
	}
	cmd := exec.Command("hostname")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(out)) != hostname {
		return fmt.Errorf("Hostname does not match, expected %s, got %s", hostname, string(out))
	}
	return nil
}
