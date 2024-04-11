package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"

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
}

func NewConfigurer(hostname string, dotfilesPath string) *Configurer {
	return &Configurer{hostname: hostname, dotfilesPath: dotfilesPath}
}

func (c *Configurer) open() (io.Reader, error) {
	configFilename := "config.toml"
	configPath := path.Join(c.dotfilesPath, "hosts", c.hostname, configFilename)
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening config file: %s", err)
	}
	fileReader := bufio.NewReader(configFile)
	return fileReader, nil
}

func (c *Configurer) parse(reader io.Reader) (ConfigugureInfo, error) {
	var config ConfigugureInfo
	// read the reader into a string
	data, err := io.ReadAll(reader)
	if err != nil {
		return config, fmt.Errorf("Error reading config file: %s", err)
	}
	buf := string(data)
	if _, err := toml.Decode(buf, &config); err != nil {
		log.Fatal(err)
		return config, err
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

func (c *Configurer) Run() error {
	reader, err := c.open()
	if err != nil {
		return err
	}
	config, err := c.parse(reader)
	if err != nil {
		return err
	}

	for _, symlinkInfo := range config.Symlinks {
		err := processSymlink(c.dotfilesPath, symlinkInfo)
		if err != nil {
			return err
		}
	}
	templater := NewTemplater(c.hostname, c.dotfilesPath)
	for _, template := range config.Templates {
		err := templater.Process(template)
		if err != nil {
			return err
		}
	}
	return nil
}
