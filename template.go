package main

import (
	"bytes"
	"log"
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
	dataMap := data.(map[string]interface{})
	t.Target = dataMap["target"].(string)
	t.Name = dataMap["name"].(string)
	t.Data = make(map[string]string)
	for k, v := range dataMap {
		if k == "target" || k == "name" {
			continue
		}
		t.Data[k] = v.(string)
	}
	return nil
}

type Templater struct {
	hostname     string
	dotfilesPath string
}

func NewTemplater(hostname string, dotfilesPath string) *Templater {
	return &Templater{hostname: hostname, dotfilesPath: dotfilesPath}
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
	if _, err := os.Stat(absTargetPath); err != nil {
		return true
	}
	lastTemplateModTime, err := os.Stat(absTemplatePath)
	if err != nil {
		return true
	}
	lastTargetModTime, err := os.Stat(absTargetPath)
	if err != nil {
		return true
	}
	return lastTemplateModTime.ModTime().After(lastTargetModTime.ModTime())
}

func (t *Templater) Process(info TemplateInfo) error {
	absTemplatePath := path.Join(t.dotfilesPath, info.Target)
	rendered, err := RenderTemplate(absTemplatePath, &info.Data)
	if err != nil {
		return err
	}
	filename := path.Base(info.Target)
	realTargetPath, err := t.getGeneratedPath(filename)
	if err != nil {
		return err
	}
	if !t.needsToUpdate(absTemplatePath, realTargetPath) {
		return nil
	}
	log.Printf("Writing rendered template to %s\n", realTargetPath)
	err = os.WriteFile(realTargetPath, []byte(rendered), os.ModePerm)
	if err != nil {
		return err
	}
	name, err := ExpandHomeDir(info.Name)
	if err != nil {
		return err
	}
	symlink := SymlinkInfo{
		Name:   name,
		Target: realTargetPath,
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
