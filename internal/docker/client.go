package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"bluem/internal/logx"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
)

type Client struct {
	Raw *docker.Client
}

func New() (*Client, error) {
	cli, err := docker.NewClientWithOpts(
		docker.FromEnv,
		docker.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	logx.L().Infow("docker client initialized",
		"host", os.Getenv("DOCKER_HOST"),
	)
	return &Client{Raw: cli}, nil
}

func (c *Client) Close() error {
	return c.Raw.Close()
}

func (c *Client) ComposeUp(ctx context.Context, file string, services ...string) error {
	args := []string{"compose", "-f", file, "up", "-d"}
	args = append(args, services...)
	logx.L().Infow("running docker compose up", "file", file, "services", services)
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) ComposeDown(ctx context.Context, file string, services ...string) error {
	args := []string{"compose", "-f", file, "down"}
	args = append(args, services...)
	logx.L().Infow("running docker compose down", "file", file, "services", services)
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) ComposePs(ctx context.Context, file string) error {
	logx.L().Infow("running docker compose ps", "file", file)
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", file, "ps")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) StopContainer(ctx context.Context, name string, timeoutSeconds int) error {
	logx.L().Infow("stopping container", "name", name, "timeout", timeoutSeconds)
	cmd := exec.CommandContext(ctx, "docker", "stop", "--time", fmt.Sprint(timeoutSeconds), name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) ContainerHealth(ctx context.Context, name string) (string, error) {
	ins, err := c.Raw.ContainerInspect(ctx, name)
	if err != nil {
		return "", err
	}
	if ins.State == nil || ins.State.Health == nil {
		return "unknown", nil
	}
	return ins.State.Health.Status, nil
}

func (c *Client) IsRunning(ctx context.Context, name string) (bool, error) {
    container, err := c.Raw.ContainerInspect(ctx, name)
    if err != nil {
        if docker.IsErrNotFound(err) {
            return false, nil
        }
        return false, err
    }

    return container.State.Running, nil
}

func (c *Client) StartContainer(ctx context.Context, name string) error {
    return c.Raw.ContainerStart(ctx, name, container.StartOptions{})
}