package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestCreateAcceptsFileReachableThroughParentSymlink(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "dotfiles", "ags", "type-shims")
	target := filepath.Join(targetDir, "binding.d.ts")
	writeFile(t, target)

	configDir := filepath.Join(tmp, ".config", "ags")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	// The whole directory is linked, so the file itself is not a symlink.
	if err := os.Symlink(targetDir, filepath.Join(configDir, "type-shims")); err != nil {
		t.Fatal(err)
	}

	link := SymlinkInfo{Name: filepath.Join(configDir, "type-shims", "binding.d.ts"), Target: target}
	if err := link.Create(); err != nil {
		t.Fatalf("file reachable through a parent symlink should be accepted, got: %s", err)
	}
}

func TestCreateRejectsUnrelatedRegularFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "dotfiles", "a.conf")
	name := filepath.Join(tmp, ".config", "a.conf")
	writeFile(t, target)
	writeFile(t, name)

	if err := (SymlinkInfo{Name: name, Target: target}).Create(); err == nil {
		t.Fatal("a regular file that is not the target must be reported, not silently accepted")
	}
}

func TestProcessSymlinkContinuesPastFailures(t *testing.T) {
	tmp := t.TempDir()
	dotfiles := filepath.Join(tmp, "dotfiles")
	writeFile(t, filepath.Join(dotfiles, "app", "bad.conf"))
	writeFile(t, filepath.Join(dotfiles, "app", "good.conf"))

	configDir := filepath.Join(tmp, ".config", "app")
	writeFile(t, filepath.Join(configDir, "bad.conf")) // conflicting regular file

	c := &Configurer{dotfilesPath: dotfiles}
	err := c.processSymlink(SymlinkInfo{Name: configDir, Target: "app"})
	if err == nil {
		t.Fatal("expected the conflicting file to be reported")
	}
	if _, statErr := os.Lstat(filepath.Join(configDir, "good.conf")); statErr != nil {
		t.Fatalf("good.conf should still have been linked despite the failure: %s", statErr)
	}
}
