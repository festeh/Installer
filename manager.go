package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
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

func getHomeDir() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %s", err)
	}
	return usr.HomeDir, nil
}

func NewManager(hostname string, opts Options) (*Manager, error) {
	if err := IsHostnameOk(hostname); err != nil {
		return nil, err
	}
	home, err := getHomeDir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(home, "dotfiles")
	ignored, _ := ReadGitignore(filepath.Join(base, ".gitignore")) // ignore error, .gitignore is optional
	return &Manager{hostname: hostname, base: base, ignored: ignored, opts: opts}, nil
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
	config, err := LoadInstallConfig(i.base, i.hostname)
	if err != nil {
		return fmt.Errorf("error loading install config: %s", err)
	}

	if len(config.Simples) == 0 {
		fmt.Println("No packages to install")
		return nil
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
	systemHostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("could not detect system hostname: %s", err)
	}
	if systemHostname != hostname {
		return fmt.Errorf("hostname mismatch: expected %s, got %s", hostname, systemHostname)
	}
	return nil
}
