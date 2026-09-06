package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type SymlinkInfo struct {
	// Example name = "~/.config/nvim/init.vim"
	Name string `toml:"name"`
	// Example target = "~/dotfiles/nvim/init.vim"
	Target string `toml:"target"`
	dryRun bool
}

func (s *SymlinkInfo) ExpandPaths(dotfilesPrefix string) error {
	absTargetPath := filepath.Join(dotfilesPrefix, s.Target)
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

func (s SymlinkInfo) checkExistingSymlink() (needsUpdate bool, err error) {
	if !s.IsTargetExists() {
		return false, fmt.Errorf("Target %s does not exist", s.Target)
	}
	// check that Name is a symlink
	fi, err := os.Lstat(s.Name)
	if err != nil {
		return false, fmt.Errorf("Error checking symlink %s: %s", s.Name, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		if s.resolvesToTarget() {
			// Reached through a symlinked parent directory
			// (e.g. ~/.config/ags/type-shims -> dotfiles/ags/type-shims), so the
			// file already lives where it should. Nothing to do.
			return false, nil
		}
		return false, fmt.Errorf("Name %s is not a symlink", s.Name)
	}
	// check if symlink points to correct target
	currentTarget, err := os.Readlink(s.Name)
	if err != nil {
		return false, fmt.Errorf("Error reading symlink %s: %s", s.Name, err)
	}
	if currentTarget != s.Target {
		return true, nil
	}
	return false, nil
}

// resolvesToTarget reports whether following every symlink in Name ends at Target.
func (s SymlinkInfo) resolvesToTarget() bool {
	resolvedName, err := filepath.EvalSymlinks(s.Name)
	if err != nil {
		return false
	}
	resolvedTarget, err := filepath.EvalSymlinks(s.Target)
	if err != nil {
		return false
	}
	return resolvedName == resolvedTarget
}

func (s SymlinkInfo) Create() error {
	// New symlink
	if _, err := os.Stat(s.Name); os.IsNotExist(err) {
		if s.dryRun {
			fmt.Printf("      would create symlink: %s\n", s.Name)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(s.Name), 0755); err != nil {
			return err
		}
		return os.Symlink(s.Target, s.Name)
	}

	// Existing - check if update needed
	needsUpdate, err := s.checkExistingSymlink()
	if err != nil {
		return err
	}
	if !needsUpdate {
		return nil
	}
	if s.dryRun {
		fmt.Printf("      would update symlink: %s\n", s.Name)
		return nil
	}
	if err := os.Remove(s.Name); err != nil {
		return fmt.Errorf("Error removing old symlink %s: %s", s.Name, err)
	}
	return os.Symlink(s.Target, s.Name)
}

func getSymlinksFromDir(symlink SymlinkInfo) ([]SymlinkInfo, error) {
	files, err := GetFiles(symlink.Target)
	symlinks := []SymlinkInfo{}
	if err != nil {
		return symlinks, err
	}
	for _, subTarget := range files {
		relPath, err := filepath.Rel(symlink.Target, subTarget)
		if err != nil {
			return symlinks, err
		}
		subName := filepath.Join(symlink.Name, relPath)
		symlinks = append(symlinks, SymlinkInfo{Name: subName, Target: subTarget, dryRun: symlink.dryRun})
	}
	return symlinks, nil
}

// processSymlink links one config entry. A directory target is linked file by
// file; every file is attempted even if some fail, and the failures are
// returned together.
func (c *Configurer) processSymlink(symlink SymlinkInfo) error {
	symlink.dryRun = c.opts.DryRun
	if err := symlink.ExpandPaths(c.dotfilesPath); err != nil {
		return err
	}
	symlinks := []SymlinkInfo{symlink}
	if isExistingDir(symlink.Target) {
		var err error
		if symlinks, err = getSymlinksFromDir(symlink); err != nil {
			return err
		}
	}
	var failures []error
	for _, link := range symlinks {
		if c.isIgnored(link) {
			continue
		}
		if err := link.Create(); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
