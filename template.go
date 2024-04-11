package main

import (
	"bytes"
	"log"
	"os"
	"path"
	"text/template"
)

type Template struct {
	Name   string `toml:"name"`
	Target string `toml:"target"`
	Data   map[string]string
}

func (t *Template) UnmarshalTOML(data interface{}) error {
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

func ProcessTemplate(t *Template, dotfilesPath string, hostname string) error {
	absTemplatePath, err := ExpandHomeDir(path.Join(dotfilesPath, t.Target))
	rendered, err := RenderTemplate(absTemplatePath, &t.Data)
	filename := path.Base(t.Target)
	absTargetPath, err := ExpandHomeDir(path.Join(dotfilesPath, "hosts", hostname, "generated", filename))
	err = os.MkdirAll(path.Dir(absTargetPath), os.ModePerm)
	if err != nil {
		return err
	}
	if _, err := os.Stat(absTemplatePath); err == nil {
		lastTemplateModTime, err := os.Stat(absTemplatePath)
		if err != nil {
			return err
		}
		lastTargetModTime, err := os.Stat(absTargetPath)
		if err != nil {
			return err
		}
		if lastTemplateModTime.ModTime().Before(lastTargetModTime.ModTime()) {
			return nil
		}
	}
	log.Printf("Writing rendered template to %s\n", absTargetPath)
	err = os.WriteFile(absTargetPath, []byte(rendered), os.ModePerm)
	if err != nil {
		return err
	}
	name, err := ExpandHomeDir(t.Name)
	if err != nil {
		return err
	}
	symlink := Symlink{
		Name:   name,
		Target: absTargetPath,
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
