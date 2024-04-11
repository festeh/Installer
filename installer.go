package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

type Installer struct {
	base     string
	hostname string
}

func NewInstaller(hostname string) *Installer {
	base := os.Getenv("HOME") + "/dotfiles"
	assertHostnameMatches(hostname)
	return &Installer{hostname: hostname, base: base}
}

func (i *Installer) Dispatch(command string) error {
	if command == "install" {
		return i.InstallFunc()
	} else if command == "config" {
		return i.ConfigFunc()
	} else {
		return fmt.Errorf("Unknown command: %s", command)
	}
}

func (i *Installer) ConfigFunc() error {

	log.Println("Configuring")
	configPath, err := GetConfigPath(i.base, i.hostname, "config.toml")
	if err != nil {
		log.Printf("Error getting config path: %s", err)
		return err
	}
	config, err := ParseConfig(configPath)
	if err != nil {
		log.Printf("Error parsing config: %s", err)
		return err
	}
	for _, symlink := range config.Symlinks {
		err := processSymlink(i.base, symlink)
		if err != nil {
			return err
		}
	}
	for _, template := range config.Templates {
		err := ProcessTemplate(i.hostname, &template)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Installer) InstallFunc() error {
	log.Println("Installing")
	// TODO: Implement
	return nil
}

// Builds a config and checks if it exists
func GetConfigPath(base string, hostname string, filename string) (string, error) {
	configPath := path.Join(base, "hosts", hostname, filename)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Error: Config %s does not exist\n", configPath)
		return "", err
	}
	return configPath, nil
}

func ExpandHomeDir(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	if path[:2] == "~/" {
		path = strings.Replace(path, "~", usr.HomeDir, 1)
	}
	return path, nil
}

func GetFiles(target string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(target, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		log.Printf("Error getting files: %s", err)
		return files, err
	}
	return files, nil
}

func isExistingDir(path string) bool {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return fi.IsDir()
}

func processSymlinkDir(symlink Symlink) error {
	files, err := GetFiles(symlink.Target)
	if err != nil {
		return err
	}
	for _, subTarget := range files {
		relPath, err := filepath.Rel(symlink.Target, subTarget)
		if err != nil {
			return err
		}
		subName := path.Join(symlink.Name, relPath)
		err = Symlink{subName, subTarget}.Create()
		if err != nil {
			return err
		}
	}
	return nil
}

func processSymlink(base string, symlink Symlink) error {
	err := symlink.ExpandPaths(base)
	if err != nil {
		return err
	}
	if isExistingDir(symlink.Target) {
		err = processSymlinkDir(symlink)
	} else {
		err = symlink.Create()
	}
	return err
}


func Exec(cmdName string, cmdArgs []string) error {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running command: %s", err)
		return err
	}
	return nil
}

func ExecCmd(cmd string, sudo bool) error {
	log.Printf("Executing command: %s\n", cmd)
	if sudo {
		cmd = "sudo " + cmd
	}
	err := Exec("bash", []string{"-c", cmd})
	if err != nil {
		log.Printf("Error executing command: %s", err)
		return err
	}
	return nil
}

// Check return status is 0
func CheckIsInstalled(check string) bool {
	cmdParts := []string{"bash", "-c", check}
	log.Printf("Checking %s command\n", cmdParts)
	cmdName := cmdParts[0]
	cmdArgs := cmdParts[1:]
	cmd := exec.Command(cmdName, cmdArgs...)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running command: %s", err)
		return false
	}
	return true
}

func InstallFunc(base string, hostname string) error {
	configPath, err := GetConfigPath(base, hostname, "install.toml")
	if err != nil {
		log.Printf("Error getting install path: %s", err)
		return err
	}
	config, err := ParseInstallConfig(configPath)
	for _, simple := range config.Simples {
		if CheckIsInstalled(simple.Check) {
			log.Printf("Skipping install of %s, already installed", simple.Check)
			continue
		}
		err = ExecCmd(simple.Cmd, simple.Sudo)
		if err != nil {
			return err
		}
		fmt.Println("")
	}
	return nil
}

func assertHostnameMatches(hostname string) {
	if hostname == "common" {
		log.Println("Skipping hostname check for common")
		return
	}
	cmd := exec.Command("hostname")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting hostname: %s", err)
		os.Exit(1)
	}
	if strings.TrimSpace(string(out)) != hostname {
		log.Printf("Error: Hostname does not match, expected %s, got %s", hostname, string(out))
		os.Exit(1)
	}
}
