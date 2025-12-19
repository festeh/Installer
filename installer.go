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
	if sudo {
		cmd = "sudo " + cmd
	}
	err := Exec("bash", []string{"-c", cmd})
	if err != nil {
		return err
	}
	return nil
}

// Check return status is 0
func CheckIsInstalled(check string) bool {
	cmdParts := []string{"bash", "-c", check}
	cmdName := cmdParts[0]
	cmdArgs := cmdParts[1:]
	cmd := exec.Command(cmdName, cmdArgs...)
	err := cmd.Run()
	return err == nil
}

