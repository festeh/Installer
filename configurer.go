package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type ConfigugureInfo struct {
	Symlinks  map[string]SymlinkInfo
	Templates map[string]TemplateInfo
}

// Function to pretty print a config
func (c ConfigugureInfo) String() string {
	res := `Symlinks: %v
Templates: %v
	`
	return fmt.Sprintf(res, c.Symlinks, c.Templates)
}

type Configurer struct {
	hostname     string
	dotfilesPath string
	ignored      []string
	opts         Options
}

func (c *Configurer) getConfigPath() string {
	return path.Join(c.dotfilesPath, "hosts", c.hostname, "config.toml")
}

func (c *Configurer) readConfig() ([]byte, error) {
	configPath := c.getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading config file: %s", err)
	}
	return data, nil
}

func (c *Configurer) parse(data []byte) (ConfigugureInfo, error) {
	var config ConfigugureInfo
	if _, err := toml.Decode(string(data), &config); err != nil {
		return config, fmt.Errorf("Error parsing config: %s", err)
	}
	for k, v := range config.Symlinks {
		if v.Name == "" {
			return config, fmt.Errorf("Symlink %s has no name", k)
		}
		if v.Target == "" {
			return config, fmt.Errorf("Symlink %s has no target", k)
		}
	}
	return config, nil
}
func (c *Configurer) isIgnored(link SymlinkInfo) bool {
	for _, ignored := range c.ignored {
		fullPath := filepath.Join(c.dotfilesPath, ignored)
		if strings.HasPrefix(link.Target, fullPath) {
			return true
		}
	}
	return false
}

func (c *Configurer) Run() error {
	data, err := c.readConfig()
	if err != nil {
		return err
	}
	config, err := c.parse(data)
	if err != nil {
		return err
	}

	symlinkCount := len(config.Symlinks)
	templateCount := len(config.Templates)
	fmt.Printf("Processing %d symlinks and %d templates...\n", symlinkCount, templateCount)

	if symlinkCount > 0 {
		fmt.Println("\n📁 Symlinks:")
	}
	for name, symlinkInfo := range config.Symlinks {
		fmt.Printf("  → %s\n", name)
		err := c.processSymlink(symlinkInfo)
		if err != nil {
			return err
		}
	}

	if templateCount > 0 {
		fmt.Println("\n📄 Templates:")
	}
	templater := NewTemplater(c.hostname, c.dotfilesPath, c.getConfigPath(), c.opts)
	for name, template := range config.Templates {
		fmt.Printf("  → %s\n", name)
		err := templater.Process(template)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Configurer) Doctor() error {
	data, err := c.readConfig()
	if err != nil {
		return err
	}
	config, err := c.parse(data)
	if err != nil {
		return err
	}

	healthy := 0
	broken := 0
	missing := 0

	fmt.Println("Checking symlinks...")
	for name, symlinkInfo := range config.Symlinks {
		err := symlinkInfo.ExpandPaths(c.dotfilesPath)
		if err != nil {
			fmt.Printf("  ✗ %s: %s\n", name, err)
			broken++
			continue
		}
		// Check if symlink exists
		fi, err := os.Lstat(symlinkInfo.Name)
		if os.IsNotExist(err) {
			fmt.Printf("  ○ %s: not created yet\n", name)
			missing++
			continue
		}
		if err != nil {
			fmt.Printf("  ✗ %s: %s\n", name, err)
			broken++
			continue
		}
		// Check if it's a symlink
		if fi.Mode()&os.ModeSymlink == 0 {
			fmt.Printf("  ✗ %s: exists but is not a symlink\n", name)
			broken++
			continue
		}
		// Check target
		currentTarget, err := os.Readlink(symlinkInfo.Name)
		if err != nil {
			fmt.Printf("  ✗ %s: cannot read symlink target\n", name)
			broken++
			continue
		}
		if currentTarget != symlinkInfo.Target {
			fmt.Printf("  ⚠ %s: points to wrong target\n", name)
			fmt.Printf("      expected: %s\n", symlinkInfo.Target)
			fmt.Printf("      actual:   %s\n", currentTarget)
			broken++
			continue
		}
		// Check if target exists
		if _, err := os.Stat(symlinkInfo.Target); os.IsNotExist(err) {
			fmt.Printf("  ✗ %s: target does not exist\n", name)
			broken++
			continue
		}
		fmt.Printf("  ✓ %s\n", name)
		healthy++
	}

	fmt.Printf("\nSummary: %d healthy, %d broken, %d missing\n", healthy, broken, missing)
	if broken > 0 {
		return fmt.Errorf("%d symlinks need attention", broken)
	}
	return nil
}

func (c *Configurer) Uninstall() error {
	data, err := c.readConfig()
	if err != nil {
		return err
	}
	config, err := c.parse(data)
	if err != nil {
		return err
	}

	removed := 0
	fmt.Println("Removing symlinks...")
	for name, symlinkInfo := range config.Symlinks {
		expandedName, err := ExpandHomeDir(symlinkInfo.Name)
		if err != nil {
			return err
		}
		// Check if it exists and is a symlink
		fi, err := os.Lstat(expandedName)
		if os.IsNotExist(err) {
			fmt.Printf("  - %s: already absent\n", name)
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking %s: %s", name, err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			fmt.Printf("  ⚠ %s: not a symlink, skipping\n", name)
			continue
		}
		if c.opts.DryRun {
			fmt.Printf("  → %s: would remove\n", name)
		} else {
			if err := os.Remove(expandedName); err != nil {
				return fmt.Errorf("error removing %s: %s", name, err)
			}
			fmt.Printf("  ✓ %s: removed\n", name)
		}
		removed++
	}

	// Also handle templates (they create symlinks too)
	for name, tmplInfo := range config.Templates {
		expandedName, err := ExpandHomeDir(tmplInfo.Name)
		if err != nil {
			return err
		}
		fi, err := os.Lstat(expandedName)
		if os.IsNotExist(err) {
			fmt.Printf("  - %s: already absent\n", name)
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking %s: %s", name, err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			fmt.Printf("  ⚠ %s: not a symlink, skipping\n", name)
			continue
		}
		if c.opts.DryRun {
			fmt.Printf("  → %s: would remove\n", name)
		} else {
			if err := os.Remove(expandedName); err != nil {
				return fmt.Errorf("error removing %s: %s", name, err)
			}
			fmt.Printf("  ✓ %s: removed\n", name)
		}
		removed++
	}

	fmt.Printf("\nRemoved %d symlinks\n", removed)
	return nil
}
