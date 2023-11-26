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

// Builds a config and checks if it exists
func GetConfigPath(base string, hostname string, filename string) (string, error) {
	configPath := path.Join(base, hostname, filename)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Error: Config %s does not exist\n", configPath)
		return "", err
	}
	return configPath, nil
}

func ExpandPath(path string) (string, error) {
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

func ConfigFunc(base string, hostname string) error {
	configPath, err := GetConfigPath(base, hostname, "config.toml")
	if err != nil {
		log.Printf("Error getting config path: %s", err)
		return err
	}
	config, err := ParseConfig(configPath)
	for _, symlink := range config.Symlinks {
		target := path.Join(base, symlink.Target)
		if _, err := os.Stat(target); os.IsNotExist(err) {
			log.Printf("Error: Target %s does not exist\n", target)
			return err
		}
		source, err := ExpandPath(symlink.Source)
		if err != nil {
			log.Printf("Error expanding path: %s", err)
			return err
		}
		if fi, err := os.Stat(target); err == nil && fi.IsDir() {
			files, err := GetFiles(target)
			if err != nil {
				return err
			}
			for _, file := range files {
				relPath, err := filepath.Rel(target, file)
				if err != nil {
					return err
				}
				sourcePath := path.Join(source, relPath)
				err = CreateSymlink(sourcePath, file)
				if err != nil {
					return err
				}
			}
		} else {
			err = CreateSymlink(source, target)

		}
		if err != nil {
			return err
		}
		fmt.Println("")
	}
	return nil
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

func main() {
	log.Printf("Starting dotfiles install")
	base := os.Getenv("HOME") + "/dotfiles"
	cmd := os.Args[1]
	hostname := os.Args[2]
	err := error(nil)
	if cmd == "install" {
		log.Printf("Running install for hostname: %s", hostname)
		err = InstallFunc(base, hostname)
	} else if cmd == "config" {
		log.Printf("Running config for hostname: %s", hostname)
		err = ConfigFunc(base, hostname)
	}
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Printf("Finished Installer")
}
