package deploy

import (
	"context"
	"fmt"
	"time"

	"bluem/internal/config"
	"bluem/internal/docker"
	"bluem/internal/logx"
	"bluem/internal/proxy"
	"bluem/internal/template"
)

type Deployer struct {
	Cfg   *config.Config
	Dock  *docker.Client
	Proxy proxy.Proxy
}

func New(ctx context.Context, configPath string) (*Deployer, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	dcli, err := docker.New()
	if err != nil {
		return nil, err
	}

	p := proxy.Get(cfg.Proxy.Type)
	if p == nil {
		return nil, fmt.Errorf("unknown proxy type %q", cfg.Proxy.Type)
	}

	logx.L().Infow("deployer initialized", "project", cfg.ProjectSlug, "proxy", p.Name())
	return &Deployer{
		Cfg:   cfg,
		Dock:  dcli,
		Proxy: p,
	}, nil
}

func (d *Deployer) Close() error {
	return d.Dock.Close()
}

func (d *Deployer) Init(ctx context.Context) error {
	logx.L().Info("generating compose & proxy configs")
	data := map[string]any{"cfg": d.Cfg}

	if err := template.RenderToFile("docker-compose.green.yml.tmpl", data, "docker-compose.green.yml"); err != nil {
		return err
	}
	if err := template.RenderToFile("docker-compose.blue.yml.tmpl", data, "docker-compose.blue.yml"); err != nil {
		return err
	}

	if err := d.Proxy.Init(d.Cfg); err != nil {
		return err
	}

	logx.L().Info("init complete")
	return nil
}

func (d *Deployer) Up(ctx context.Context) error {
	logx.L().Info("bringing up GREEN stack")
	if err := d.Dock.ComposeUp(ctx, "docker-compose.green.yml"); err != nil {
		return err
	}
	logx.L().Infow("bringing up proxy", "proxy", d.Proxy.Name())
	if err := d.Proxy.Up(ctx, d.Dock); err != nil {
		return err
	}
	return nil
}

func (d *Deployer) waitHealthy(ctx context.Context, container string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := d.Dock.ContainerHealth(ctx, container)
		if err != nil {
			return err
		}
		logx.L().Debugw("container health", "container", container, "status", status)
		if status == "healthy" {
			return nil
		}
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("container %s did not become healthy in time", container)
}

func (d *Deployer) Deploy(ctx context.Context) error {
    logx.L().Info("=== BLUE-GREEN DEPLOY START ===")

    // 1. Guarantee GREEN is running
    if err := d.EnsureGreenActive(ctx); err != nil {
        return fmt.Errorf("failed to ensure green active: %w", err)
    }

    // 2. Start BLUE stack
    if err := d.Dock.ComposeUp(ctx, "docker-compose.blue.yml"); err != nil {
        return err
    }

    blueName := d.Cfg.ProjectSlug + "_blue"
    if err := d.waitHealthy(ctx, blueName, 2*time.Minute); err != nil {
        return err
    }

    // 3. Switch Proxy to BLUE
    if err := d.Proxy.SwitchToBlue(ctx, d.Dock, d.Cfg); err != nil {
        return err
    }

    // 4. Auto-Failover Monitor
    if d.Cfg.Deployment.AutoRollback {
        if err := d.AutoFailoverMonitor(ctx); err != nil {
            return err
        }
    }

    // 5. Drain GREEN (traffic already gone)
    logx.L().Infow("draining GREEN", "seconds",  d.Cfg.ContainerDrainingTimeout)
    time.Sleep(d.Cfg.ContainerDrainingTimeout)

    // 6. Stop GREEN
    greenName := d.Cfg.ProjectSlug + "_green"

    logx.L().Infow("stopping GREEN container", "name", greenName)
	timeout := int(d.Cfg.ContainerDrainingTimeout.Seconds())
    if err := d.Dock.StopContainer(ctx, greenName, timeout); err != nil {
        logx.L().Warnw("failed to stop green", "error", err)
    }

    logx.L().Info("=== BLUE-GREEN DEPLOY COMPLETE ===")
    return nil
}

func (d *Deployer) Rollback(ctx context.Context) error {
	logx.L().Info("=== ROLLBACK START ===")
	if err := d.EnsureGreenActive(ctx); err != nil {
        return fmt.Errorf("failed to ensure green active: %w", err)
    }
	if err := d.Proxy.SwitchToGreen(ctx, d.Dock, d.Cfg); err != nil {
		return err
	}
	blueName := d.Cfg.ProjectSlug + "_blue"
	timeout := int(d.Cfg.ContainerDrainingTimeout.Seconds())
	if err := d.Dock.StopContainer(ctx, blueName, timeout); err != nil {
        logx.L().Warnw("failed to stop blue", "error", err)
    }

	logx.L().Info("=== ROLLBACK COMPLETE ===")
	return nil
}

func (d *Deployer) Status(ctx context.Context) error {
	logx.L().Info("status: GREEN/BLUE stacks")
	if err := d.Dock.ComposePs(ctx, "docker-compose.green.yml"); err != nil {
		return err
	}
	_ = d.Dock.ComposePs(ctx, "docker-compose.blue.yml")

	logx.L().Infow("status: proxy", "proxy", d.Proxy.Name())
	return d.Proxy.Status(ctx, d.Dock)
}

func (d *Deployer) EnsureGreenActive(ctx context.Context) error {
    greenName := d.Cfg.ProjectSlug + "_green"

    running, err := d.Dock.IsRunning(ctx, greenName)
    if err != nil {
        return err
    }

    if running {
        logx.L().Info("GREEN container is already running")
        return nil
    }

    logx.L().Warn("GREEN is not running — starting it now")
    if err := d.Dock.StartContainer(ctx, greenName); err != nil {
        return err
    }

    logx.L().Info("GREEN container started successfully")
    return nil
}


func (d *Deployer) AutoFailoverMonitor(ctx context.Context) error {
    blue := d.Cfg.ProjectSlug + "_blue"

    monitorFor := d.Cfg.FaileoverTimeout
    checkEvery := d.Cfg.HealthCheckDelay
    timeout := time.Now().Add(monitorFor)

    logx.L().Infow("Monitoring BLUE for auto-failover", "duration", monitorFor)

    for time.Now().Before(timeout) {
		status, err := d.Dock.ContainerHealth(ctx, blue)
        if status != "healthy" || err != nil {
            logx.L().Errorw("BLUE became unhealthy! Auto-failover triggered", "error", err)

            // Switch TRAFFIC back to GREEN
            if err := d.Proxy.SwitchToGreen(ctx, d.Dock, d.Cfg); err != nil {
                return fmt.Errorf("auto-failover failed: %w", err)
            }

            logx.L().Warn("Traffic switched back to GREEN")

            return fmt.Errorf("BLUE unhealthy — failover executed")
        }

        time.Sleep(checkEvery)
    }

    logx.L().Info("BLUE remained healthy — auto-failover completed successfully")
    return nil
}


func (d *Deployer) Stop(ctx context.Context) error {
	logx.L().Info("=== STOPPING ALL STACKS & PROXY ===")

	_ = d.Dock.ComposeDown(ctx, "docker-compose.green.yml")
	_ = d.Dock.ComposeDown(ctx, "docker-compose.blue.yml")
	_ = d.Proxy.Stop(ctx, d.Dock, d.Cfg)

	logx.L().Info("All stacks stopped.")
	return nil
}