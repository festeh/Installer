package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	base     string
	hostname string
	ignored  []string
}

func ReadGitignore(path string) ([]string, error) {
	gitignore, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer gitignore.Close()
	scanner := bufio.NewScanner(gitignore)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func NewManager(hostname string) (*Manager, error) {
	if err := IsHostnameOk(hostname); err != nil {
		return nil, err
	}
	base := filepath.Join(os.Getenv("HOME"), "dotfiles")
	expanded, err := ExpandHomeDir(base)
	if err != nil {
		return nil, err
	}
	ignored, err := ReadGitignore(filepath.Join(expanded, ".gitignore"))
	return &Manager{hostname: hostname, base: expanded, ignored: ignored}, nil
}

func (i *Manager) Dispatch(command string) error {
	if command == "install" {
		// TODO: Implement install
		return nil
	} else if command == "config" {
		configurer := &Configurer{hostname: i.hostname, dotfilesPath: i.base, ignored: i.ignored}
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
