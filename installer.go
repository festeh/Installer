package main

import (
	"log"
	"os"
	"os/exec"
)

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

// func InstallFunc(base string, hostname string) error {
// 	configPath, err := GetConfigPath(base, hostname, "install.toml")
// 	if err != nil {
// 		log.Printf("Error getting install path: %s", err)
// 		return err
// 	}
// 	config, err := ParseInstallConfig(configPath)
// 	for _, simple := range config.Simples {
// 		if CheckIsInstalled(simple.Check) {
// 			log.Printf("Skipping install of %s, already installed", simple.Check)
// 			continue
// 		}
// 		err = ExecCmd(simple.Cmd, simple.Sudo)
// 		if err != nil {
// 			return err
// 		}
// 		fmt.Println("")
// 	}
// 	return nil
// }
