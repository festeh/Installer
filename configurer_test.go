package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesGitignore(t *testing.T) {
	patterns := []string{"node_modules", "ags/@girs", "/.aider*", "*.py[cod]", "spotify/.venv/", "ags/types"}
	cases := []struct {
		rel  string
		want bool
	}{
		{"ags/node_modules/@girs/atk-1.0/README.md", true}, // bare name matches at any depth
		{"node_modules/x.js", true},
		{"ags/@girs/gtk.d.ts", true}, // anchored directory prefix
		{"ags/types/foo.d.ts", true},
		{".aider.conf.yml", true}, // leading slash anchors a glob to the root
		{"nvim/.aider.conf.yml", false},
		{"scripts/__pycache__/x.pyc", true}, // glob on a name component
		{"scripts/x.py", false},
		{"spotify/.venv/bin/python", true}, // trailing-slash directory pattern
		{"ags/widget/Menu.tsx", false},
		{"hyprland/hyprland_template.lua", false},
	}
	for _, tc := range cases {
		if got := matchesGitignore(patterns, tc.rel); got != tc.want {
			t.Errorf("matchesGitignore(%q) = %v, want %v", tc.rel, got, tc.want)
		}
	}
}

func TestIsIgnoredOnlyInsideDotfiles(t *testing.T) {
	c := &Configurer{dotfilesPath: "/home/u/dotfiles", ignored: []string{"node_modules"}}
	if c.isIgnored(SymlinkInfo{Target: "/home/u/other/node_modules/x"}) {
		t.Errorf("paths outside the dotfiles root must never be ignored")
	}
	if !c.isIgnored(SymlinkInfo{Target: "/home/u/dotfiles/ags/node_modules/x"}) {
		t.Errorf("expected ags/node_modules/x to be ignored")
	}
}

func TestReadGitignoreSkipsNoise(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".gitignore")
	content := "# comment\n\nnode_modules\n  ags/@girs  \n!keep.me\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadGitignore(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"node_modules", "ags/@girs"}
	if len(got) != len(want) {
		t.Fatalf("ReadGitignore = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ReadGitignore[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
