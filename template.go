package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"text/template"
)

type TemplateInfo struct {
	Name   string `toml:"name"`
	Target string `toml:"target"`
	Data   map[string]string
}

func (t *TemplateInfo) UnmarshalTOML(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid template config: expected map")
	}
	target, ok := dataMap["target"].(string)
	if !ok {
		return fmt.Errorf("template missing 'target' field or not a string")
	}
	t.Target = target
	name, ok := dataMap["name"].(string)
	if !ok {
		return fmt.Errorf("template missing 'name' field or not a string")
	}
	t.Name = name
	t.Data = make(map[string]string)
	for k, v := range dataMap {
		if k == "target" || k == "name" {
			continue
		}
		strVal, ok := v.(string)
		if !ok {
			return fmt.Errorf("template variable '%s' must be a string", k)
		}
		t.Data[k] = strVal
	}
	return nil
}

type Templater struct {
	hostname     string
	dotfilesPath string
	configPath   string
	opts         Options
}

func NewTemplater(hostname string, dotfilesPath string, configPath string, opts Options) *Templater {
	return &Templater{hostname: hostname, dotfilesPath: dotfilesPath, configPath: configPath, opts: opts}
}

func (t *Templater) getGeneratedPath(filename string) (string, error) {
	absTargetPath := path.Join(t.dotfilesPath, "hosts", t.hostname, "generated", filename)
	err := os.MkdirAll(path.Dir(absTargetPath), os.ModePerm)
	if err != nil {
		return "", err
	}
	return absTargetPath, nil
}

func (t *Templater) needsToUpdate(absTemplatePath string, absTargetPath string) bool {
	targetStat, err := os.Stat(absTargetPath)
	if err != nil {
		return true
	}
	targetMtime := targetStat.ModTime()

	// Check if template file is newer
	templateStat, err := os.Stat(absTemplatePath)
	if err != nil {
		return true
	}
	if templateStat.ModTime().After(targetMtime) {
		return true
	}

	// Check if config.toml is newer (template data may have changed)
	configStat, err := os.Stat(t.configPath)
	if err != nil {
		return true
	}
	return configStat.ModTime().After(targetMtime)
}

func (t *Templater) Process(info TemplateInfo) error {
	absTemplatePath := path.Join(t.dotfilesPath, info.Target)
	filename := path.Base(info.Target)
	realTargetPath, err := t.getGeneratedPath(filename)
	if err != nil {
		return err
	}

	// Only render and write if template or config changed (or --force)
	needsUpdate := t.opts.Force || t.needsToUpdate(absTemplatePath, realTargetPath)
	if needsUpdate {
		if t.opts.DryRun {
			fmt.Printf("      would regenerate %s\n", filename)
		} else {
			rendered, err := RenderTemplate(absTemplatePath, &info.Data)
			if err != nil {
				return err
			}
			err = os.WriteFile(realTargetPath, []byte(rendered), os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	name, err := ExpandHomeDir(info.Name)
	if err != nil {
		return err
	}
	symlink := SymlinkInfo{
		Name:   name,
		Target: realTargetPath,
		dryRun: t.opts.DryRun,
	}
	return symlink.Create()
}

func RenderTemplate(templatePath string, data *map[string]string) (string, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}
