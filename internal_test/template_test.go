package internal_test

import (
	"testing"

	"bluem/internal/config"
	"bluem/internal/template"
)

func TestTemplateRender(t *testing.T) {
	cfg := &config.Config{
		ProjectSlug: "demo",
		AppPort:     8000,
		GreenPort:   8010,
		BluePort:    8011,
		NetworkName: "app-network",
	}

	data := map[string]any{"cfg": cfg}
	s, err := template.Render("docker-compose.green.yml.tmpl", data)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) == 0 {
		t.Fatal("rendered template is empty")
	}
}