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
	Inherits  string `toml:"inherits"`
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
	return c.getConfigPathForHost(c.hostname)
}

func (c *Configurer) getConfigPathForHost(host string) string {
	return path.Join(c.dotfilesPath, "hosts", host, "config.toml")
}

func (c *Configurer) listAvailableHosts() []string {
	hostsDir := path.Join(c.dotfilesPath, "hosts")
	entries, err := os.ReadDir(hostsDir)
	if err != nil {
		return nil
	}
	var hosts []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		configPath := path.Join(hostsDir, entry.Name(), "config.toml")
		if _, err := os.Stat(configPath); err == nil {
			hosts = append(hosts, entry.Name())
		}
	}
	return hosts
}

func (c *Configurer) readConfig() ([]byte, error) {
	configPath := c.getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			hosts := c.listAvailableHosts()
			if len(hosts) > 0 {
				return nil, fmt.Errorf("no config found for host '%s'\nAvailable hosts: %v", c.hostname, hosts)
			}
			return nil, fmt.Errorf("no config found for host '%s' (no hosts configured)", c.hostname)
		}
		return nil, fmt.Errorf("error reading config file: %s", err)
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

func (c *Configurer) loadConfig() (ConfigugureInfo, error) {
	return c.loadConfigForHost(c.hostname, make(map[string]bool))
}

func (c *Configurer) loadConfigForHost(host string, visited map[string]bool) (ConfigugureInfo, error) {
	if visited[host] {
		return ConfigugureInfo{}, fmt.Errorf("circular inheritance detected: %s", host)
	}
	visited[host] = true

	configPath := c.getConfigPathForHost(host)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ConfigugureInfo{}, fmt.Errorf("inherited config '%s' not found", host)
		}
		return ConfigugureInfo{}, fmt.Errorf("error reading config: %s", err)
	}

	config, err := c.parse(data)
	if err != nil {
		return config, err
	}

	if config.Inherits == "" {
		return config, nil
	}

	parent, err := c.loadConfigForHost(config.Inherits, visited)
	if err != nil {
		return config, fmt.Errorf("error loading parent '%s': %s", config.Inherits, err)
	}

	return mergeConfigs(parent, config), nil
}

func mergeConfigs(parent, child ConfigugureInfo) ConfigugureInfo {
	result := ConfigugureInfo{
		Symlinks:  make(map[string]SymlinkInfo),
		Templates: make(map[string]TemplateInfo),
	}
	for k, v := range parent.Symlinks {
		result.Symlinks[k] = v
	}
	for k, v := range parent.Templates {
		result.Templates[k] = v
	}
	for k, v := range child.Symlinks {
		result.Symlinks[k] = v
	}
	for k, v := range child.Templates {
		result.Templates[k] = v
	}
	return result
}

// isIgnored reports whether a symlink target is covered by one of the
// .gitignore patterns of the dotfiles repository.
func (c *Configurer) isIgnored(link SymlinkInfo) bool {
	rel, err := filepath.Rel(c.dotfilesPath, link.Target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return false
	}
	return matchesGitignore(c.ignored, rel)
}

// matchesGitignore applies the common gitignore rules to a path relative to
// the repository root: a pattern without a slash matches a file or directory
// name at any depth ("node_modules" hides ags/node_modules/...), a pattern
// with a slash is anchored to the root ("ags/types", "/.aider*"), and a
// trailing slash is accepted as a directory marker. Globs use filepath.Match.
func matchesGitignore(patterns []string, rel string) bool {
	components := strings.Split(filepath.ToSlash(rel), "/")
	for _, pattern := range patterns {
		pattern = strings.TrimSuffix(pattern, "/")
		if pattern == "" {
			continue
		}
		if !strings.Contains(pattern, "/") {
			for _, component := range components {
				if ok, _ := filepath.Match(pattern, component); ok {
					return true
				}
			}
			continue
		}
		pattern = strings.TrimPrefix(pattern, "/")
		for i := 1; i <= len(components); i++ {
			if ok, _ := filepath.Match(pattern, strings.Join(components[:i], "/")); ok {
				return true
			}
		}
	}
	return false
}

func (c *Configurer) Run() error {
	config, err := c.loadConfig()
	if err != nil {
		return err
	}

	symlinkCount := len(config.Symlinks)
	templateCount := len(config.Templates)
	fmt.Printf("Processing %d symlinks and %d templates...\n", symlinkCount, templateCount)

	// A single broken entry must not stop the rest of the run: report it,
	// keep going, and fail at the end so the exit code still reflects it.
	var failed []string

	if symlinkCount > 0 {
		fmt.Println("\n📁 Symlinks:")
	}
	for name, symlinkInfo := range config.Symlinks {
		fmt.Printf("  → %s\n", name)
		if err := c.processSymlink(symlinkInfo); err != nil {
			fmt.Printf("  ✗ %s: %s\n", name, err)
			failed = append(failed, "symlink "+name)
		}
	}

	if templateCount > 0 {
		fmt.Println("\n📄 Templates:")
	}
	templater := NewTemplater(c.hostname, c.dotfilesPath, c.getConfigPath(), c.opts)
	for name, template := range config.Templates {
		fmt.Printf("  → %s\n", name)
		if err := templater.Process(template); err != nil {
			fmt.Printf("  ✗ %s: %s\n", name, err)
			failed = append(failed, "template "+name)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("%d of %d entries failed: %s", len(failed), symlinkCount+templateCount, strings.Join(failed, ", "))
	}
	return nil
}

func (c *Configurer) Doctor() error {
	config, err := c.loadConfig()
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

func (c *Configurer) removeSymlink(name, path string) (removed bool, err error) {
	expandedPath, err := ExpandHomeDir(path)
	if err != nil {
		return false, err
	}
	fi, err := os.Lstat(expandedPath)
	if os.IsNotExist(err) {
		fmt.Printf("  - %s: already absent\n", name)
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error checking %s: %s", name, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		fmt.Printf("  ⚠ %s: not a symlink, skipping\n", name)
		return false, nil
	}
	if c.opts.DryRun {
		fmt.Printf("  → %s: would remove\n", name)
		return true, nil
	}
	if err := os.Remove(expandedPath); err != nil {
		return false, fmt.Errorf("error removing %s: %s", name, err)
	}
	fmt.Printf("  ✓ %s: removed\n", name)
	return true, nil
}

func (c *Configurer) Uninstall() error {
	config, err := c.loadConfig()
	if err != nil {
		return err
	}

	removed := 0
	fmt.Println("Removing symlinks...")
	for name, symlinkInfo := range config.Symlinks {
		ok, err := c.removeSymlink(name, symlinkInfo.Name)
		if err != nil {
			return err
		}
		if ok {
			removed++
		}
	}
	for name, tmplInfo := range config.Templates {
		ok, err := c.removeSymlink(name, tmplInfo.Name)
		if err != nil {
			return err
		}
		if ok {
			removed++
		}
	}

	fmt.Printf("\nRemoved %d symlinks\n", removed)
	return nil
}
