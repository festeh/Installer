package main

import (
	"bytes"
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

func RenderTemplate(templatePath string, data *map[string]string) (string, error) {
	tmpl, err := template.New("config").ParseFiles(templatePath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}
