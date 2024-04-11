package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

type Symlink struct {
	Name   string `toml:"name"`
	Target string `toml:"target"`
}

func (s Symlink) ExpandPaths(dotfilesPrefix string) error {
	absTargetPath, err := ExpandHomeDir(path.Join(dotfilesPrefix, s.Target))
	if err != nil {
		return err
	}
	absNamePath, err := ExpandHomeDir(s.Name)
	if err != nil {
		return err
	}
	s.Name = absNamePath
	s.Target = absTargetPath
	if !s.IsTargetExists() {
		return fmt.Errorf("Broken symlink: Target %s does not exist", s.Target)
	}
	return nil
}

func (s Symlink) IsTargetExists() bool {
	_, err := os.Stat(s.Target)
	return !os.IsNotExist(err)
}

func (s Symlink) checkExistingSymlink() error {
	if !s.IsTargetExists() {
		return fmt.Errorf("Target %s does not exist", s.Target)
	}
	// check that Name is a symlink
	fi, err := os.Lstat(s.Name)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("Name %s is not a symlink", s.Name)
	}
	return nil
}

func (s Symlink) Create() error {
	if _, err := os.Stat(s.Name); !os.IsNotExist(err) {
		err := s.checkExistingSymlink()
		if err != nil {
			return err
		}
		return nil
	}
	log.Printf("Creating symlink %s -> %s\n", s.Name, s.Target)
	err := os.MkdirAll(path.Dir(s.Name), 0755)
	if err != nil {
		return err
	}
	return os.Symlink(s.Target, s.Name)
}
