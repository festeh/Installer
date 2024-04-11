package main

import "testing"

func TestRenderTemplate(t *testing.T) {
	data := map[string]string{
		"bar": "foo",
	}
	tmpl, err := RenderTemplate("test/test_template.tmpl", &data)
	if err != nil {
		t.Errorf("Error rendering template: %s", err)
	}
	if tmpl != "foofoo bazbaz\n" {
		t.Errorf("Expected 'foofoo bazbaz', got '%s'", tmpl)
	}
}
