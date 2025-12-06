package template

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/* templates/nginx/* templates/traefik/*
var TFS embed.FS

func Render(name string, data any) (string, error) {
	// e.g. name = "docker-compose.green.yml.tmpl"
	path := "templates/" + name

	t, err := template.ParseFS(TFS, path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderToFile(name string, data any, outPath string) error {
	out, err := Render(name, data)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(outPath, []byte(out), 0o644)
}