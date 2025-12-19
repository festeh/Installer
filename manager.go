package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	DryRun bool
	Force  bool
}

type Manager struct {
	base     string
	hostname string
	ignored  []string
	opts     Options
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

func NewManager(hostname string, opts Options) (*Manager, error) {
	if err := IsHostnameOk(hostname); err != nil {
		return nil, err
	}
	base := filepath.Join(os.Getenv("HOME"), "dotfiles")
	expanded, err := ExpandHomeDir(base)
	if err != nil {
		return nil, err
	}
	ignored, err := ReadGitignore(filepath.Join(expanded, ".gitignore"))
	return &Manager{hostname: hostname, base: expanded, ignored: ignored, opts: opts}, nil
}

func (i *Manager) Dispatch(command string) error {
	if i.opts.DryRun {
		fmt.Println("🔍 DRY RUN MODE - no changes will be made\n")
	}
	switch command {
	case "install":
		fmt.Printf("\n📦 Running install for host: %s\n\n", i.hostname)
		return i.runInstall()
	case "config":
		fmt.Printf("\n⚙️  Running config for host: %s\n\n", i.hostname)
		configurer := &Configurer{hostname: i.hostname, dotfilesPath: i.base, ignored: i.ignored, opts: i.opts}
		return configurer.Run()
	case "doctor":
		fmt.Printf("\n🩺 Running doctor for host: %s\n\n", i.hostname)
		configurer := &Configurer{hostname: i.hostname, dotfilesPath: i.base, ignored: i.ignored, opts: i.opts}
		return configurer.Doctor()
	case "uninstall":
		fmt.Printf("\n🗑️  Running uninstall for host: %s\n\n", i.hostname)
		configurer := &Configurer{hostname: i.hostname, dotfilesPath: i.base, ignored: i.ignored, opts: i.opts}
		return configurer.Uninstall()
	default:
		return fmt.Errorf("Unknown command: %s. Available: config, install, doctor, uninstall", command)
	}
}

func (i *Manager) runInstall() error {
	configPath := filepath.Join(i.base, "hosts", i.hostname, "install.toml")
	fmt.Printf("Reading install config: %s\n", configPath)
	config, err := ParseInstallConfig(configPath)
	if err != nil {
		return fmt.Errorf("Error parsing install config: %s", err)
	}

	if len(config.Simples) == 0 {
		return fmt.Errorf("No packages to install in %s", configPath)
	}

	for name, simple := range config.Simples {
		if CheckIsInstalled(simple.Check) {
			fmt.Printf("✓ Skipping %s (already installed)\n", name)
			continue
		}
		fmt.Printf("→ Installing %s...\n", name)
		err = ExecCmd(simple.Cmd, simple.Sudo)
		if err != nil {
			return fmt.Errorf("Error installing %s: %s", name, err)
		}
		fmt.Printf("✓ Installed %s successfully\n", name)
	}
	return nil
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
