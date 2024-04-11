package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
)

type SymlinkInfo struct {
	Name   string `toml:"name"`
	Target string `toml:"target"`
}

func (s *SymlinkInfo) ExpandPaths(dotfilesPrefix string) error {
	absTargetPath := path.Join(dotfilesPrefix, s.Target)
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

func (s SymlinkInfo) IsTargetExists() bool {
	_, err := os.Stat(s.Target)
	return !os.IsNotExist(err)
}

func (s SymlinkInfo) checkExistingSymlink() error {
	if !s.IsTargetExists() {
		return fmt.Errorf("Target %s does not exist", s.Target)
	}
	// check that Name is a symlink
	fi, err := os.Lstat(s.Name)
	if err != nil {
		return fmt.Errorf("Error checking symlink %s: %s", s.Name, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("Name %s is not a symlink", s.Name)
	}
	return nil
}

func (s SymlinkInfo) Create() error {
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

func processSymlinkDir(symlink SymlinkInfo) error {
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
		err = SymlinkInfo{subName, subTarget}.Create()
		if err != nil {
			return err
		}
	}
	return nil
}

func processSymlink(dotfilesPath string, symlink SymlinkInfo) error {
	err := symlink.ExpandPaths(dotfilesPath)
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
