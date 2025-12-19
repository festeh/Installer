package main

import (
	"os/user"
	"strings"
)

func ExpandHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return strings.Replace(path, "~", usr.HomeDir, 1), nil
}
