package proxy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"bluem/internal/config"
	"bluem/internal/docker"
	"bluem/internal/logx"
	"bluem/internal/template"
)

type TraefikProxy struct{}

func (TraefikProxy) Name() string { return "traefik" }

func (TraefikProxy) Init(cfg *config.Config) error {
	ctx := map[string]any{"cfg": cfg}

	if err := template.RenderToFile("docker-compose.traefik.yml.tmpl", ctx, "docker-compose.traefik.yml"); err != nil {
		return err
	}
	if err := template.RenderToFile("traefik/traefik.yml.tmpl", ctx, "traefik/traefik.yml"); err != nil {
		return err
	}
	dynPath := filepath.Join("traefik", "dynamic", fmt.Sprintf("%s.yml", cfg.ProjectSlug))
	if err := os.MkdirAll(filepath.Dir(dynPath), 0o755); err != nil {
		return err
	}
	if err := template.RenderToFile("traefik/dynamic.yml.tmpl", ctx, dynPath); err != nil {
		return err
	}
	logx.L().Infow("traefik proxy files generated")
	return nil
}

func (TraefikProxy) Up(ctx context.Context, cli *docker.Client) error {
	logx.L().Infow("starting traefik proxy")
	return cli.ComposeUp(ctx, "docker-compose.traefik.yml", "traefik")
}

func sedInPlace(path, from, to string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("sed", "-i", "", fmt.Sprintf("s/%s/%s/", from, to), path)
	} else {
		cmd = exec.Command("sed", "-i", fmt.Sprintf("s/%s/%s/", from, to), path)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (TraefikProxy) SwitchToBlue(ctx context.Context, cli *docker.Client, cfg *config.Config) error {
	file := filepath.Join("traefik", "dynamic", fmt.Sprintf("%s.yml", cfg.ProjectSlug))
	logx.L().Infow("switching traefik to BLUE", "file", file)
	if err := sedInPlace(file,
		fmt.Sprintf("service: %s_green", cfg.ProjectSlug),
		fmt.Sprintf("service: %s_blue", cfg.ProjectSlug),
	); err != nil {
		return err
	}
	_ = exec.Command("docker", "kill", "-s", "HUP", "traefik").Run()
	logx.L().Infow("switched traefik to BLUE")
	return nil
}

func (TraefikProxy) SwitchToGreen(ctx context.Context, cli *docker.Client, cfg *config.Config) error {
	file := filepath.Join("traefik", "dynamic", fmt.Sprintf("%s.yml", cfg.ProjectSlug))
	logx.L().Infow("switching traefik to GREEN", "file", file)
	if err := sedInPlace(file,
		fmt.Sprintf("service: %s_blue", cfg.ProjectSlug),
		fmt.Sprintf("service: %s_green", cfg.ProjectSlug),
	); err != nil {
		return err
	}
	_ = exec.Command("docker", "kill", "-s", "HUP", "traefik").Run()
	return nil
}

func (TraefikProxy) Status(ctx context.Context, cli *docker.Client) error {
	return cli.ComposePs(ctx, "docker-compose.traefik.yml")
}

func (p *TraefikProxy) Stop(ctx context.Context, cli *docker.Client, cfg *config.Config) error {
    return cli.ComposeDown(ctx, "docker-compose.traefik.yml")
}

func init() {
	_ = os.MkdirAll("traefik/dynamic", 0o755)
	Register(&TraefikProxy{})
}