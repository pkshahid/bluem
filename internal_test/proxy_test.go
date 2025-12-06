package internal_test

import (
	"testing"

	"bluem/internal/proxy"
)

func TestProxyRegistry(t *testing.T) {
	p := proxy.Get("nginx")
	if p == nil {
		t.Fatal("nginx proxy not registered")
	}
	if p.Name() != "nginx" {
		t.Fatalf("expected name nginx, got %s", p.Name())
	}

	p2 := proxy.Get("traefik")
	if p2 == nil {
		t.Fatal("traefik proxy not registered")
	}
}