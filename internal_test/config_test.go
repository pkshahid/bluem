package internal_test

import (
	"os"
	"testing"

	"bluem/internal/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	yaml := `
project_name: test
project_slug: demo
registry: local
image: web
`

	if err := os.WriteFile("test_cfg.yml", []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_cfg.yml")

	cfg, err := config.Load("test_cfg.yml")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ProjectSlug != "demo" {
		t.Fatalf("expected demo, got %s", cfg.ProjectSlug)
	}
	if cfg.Proxy.Type != "nginx" {
		t.Fatalf("expected default proxy nginx, got %s", cfg.Proxy.Type)
	}
}