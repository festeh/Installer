package main

import (
	"os/user"
	"strings"
)

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
