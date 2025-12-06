package proxy

import (
	"context"

	"bluem/internal/config"
	"bluem/internal/docker"
)

type Proxy interface {
	Name() string
	Init(cfg *config.Config) error
	Up(ctx context.Context, cli *docker.Client) error
	SwitchToBlue(ctx context.Context, cli *docker.Client, cfg *config.Config) error
	SwitchToGreen(ctx context.Context, cli *docker.Client, cfg *config.Config) error
	Status(ctx context.Context, cli *docker.Client) error
	Stop(ctx context.Context, cli *docker.Client, cfg *config.Config) error
}

var registry = map[string]Proxy{}

func Register(p Proxy) {
	registry[p.Name()] = p
}

func Get(name string) Proxy {
	if p, ok := registry[name]; ok {
		return p
	}
	return nil
}

func All() map[string]Proxy {
	return registry
}