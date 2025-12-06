package proxy

import (
	"context"

	"bluem/internal/config"
	"bluem/internal/docker"
	"bluem/internal/logx"
	"bluem/internal/template"
)

type NginxProxy struct{}

func (NginxProxy) Name() string { return "nginx" }

func (NginxProxy) Init(cfg *config.Config) error {
	ctx := map[string]any{"cfg": cfg}

	if err := template.RenderToFile("nginx/nginx.conf.tmpl", ctx, "nginx/nginx.conf"); err != nil {
		return err
	}
	if err := template.RenderToFile("nginx/default.conf.template.tmpl", ctx, "nginx/default.conf.template"); err != nil {
		return err
	}
	if err := template.RenderToFile("docker-compose.nginx.yml.tmpl", ctx, "docker-compose.nginx.yml"); err != nil {
		return err
	}
	logx.L().Infow("nginx proxy files generated")
	return nil
}

func (NginxProxy) Up(ctx context.Context, cli *docker.Client) error {
	logx.L().Infow("starting nginx proxy")
	return cli.ComposeUp(ctx, "docker-compose.nginx.yml", "nginx")
}

func (NginxProxy) SwitchToBlue(ctx context.Context, cli *docker.Client, cfg *config.Config) error {
	logx.L().Infow("switching nginx to BLUE", "host", cfg.ProjectSlug+"_blue")
	// In a more advanced version, you'd modify ACTIVE_APP_HOST via env/compose override.
	return cli.ComposeUp(ctx, "docker-compose.nginx.yml", "nginx")
}

func (NginxProxy) SwitchToGreen(ctx context.Context, cli *docker.Client, cfg *config.Config) error {
	logx.L().Infow("switching nginx to GREEN", "host", cfg.ProjectSlug+"_green")
	return cli.ComposeUp(ctx, "docker-compose.nginx.yml", "nginx")
}

func (NginxProxy) Status(ctx context.Context, cli *docker.Client) error {
	return cli.ComposePs(ctx, "docker-compose.nginx.yml")
}

func (NginxProxy) Stop(ctx context.Context, cli *docker.Client, cfg *config.Config) error {
    return cli.ComposeDown(ctx, "docker-compose.nginx.yml")
}

func init() {
	Register(&NginxProxy{})
}