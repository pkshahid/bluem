package config

import (
	"os"
	"time"
	"gopkg.in/yaml.v3"
)


type DeploymentConfig struct {
    AutoRollback          bool `yaml:"auto_rollback"`
    MonitorSeconds        int  `yaml:"monitor_seconds"`
    CheckIntervalSeconds  int  `yaml:"check_interval_seconds"`
}

type ProxyConfig struct {
	Type string `yaml:"type"` // "nginx", "traefik", or custom
	Port int    `yaml:"port"` // external HTTP port
}

type RemoteConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`    // "deploy@1.2.3.4:22"
	Key     string `yaml:"key"`     // "~/.ssh/id_rsa"
	WorkDir string `yaml:"workdir"` // "/home/deploy/myapp"
}

type AppConfig struct {
	Enabled      bool   `yaml:"enabled"`
	EnvFile      string `yaml:"env_file"`
	StaticVolume string `yaml:"static_volume"`
	MediaVolume  string `yaml:"media_volume"`
}

type ServicesConfig struct {
	App AppConfig `yaml:"app"`
}


type Config struct {
	ProjectName string `yaml:"project_name"`
	ProjectSlug string `yaml:"project_slug"`

	Registry string `yaml:"registry"`
	Image    string `yaml:"image"`
	Tag      string `yaml:"tag"`

	AppPort    int    `yaml:"app_port"`
	GreenPort  int    `yaml:"green_port"`
	BluePort   int    `yaml:"blue_port"`
	HealthPath string `yaml:"health_path"`

	NetworkName     string `yaml:"network_name"`
	NetworkExternal bool   `yaml:"network_external"`

	FaileoverTimeout time.Duration	`yaml:"failover_monitor_timeout"`
	ContainerDrainingTimeout time.Duration	`yaml:"container_drain_timeout"`
	HealthCheckDelay time.Duration	`yaml:"health_check_delay"`

	Proxy    ProxyConfig    `yaml:"proxy"`
	Remote   RemoteConfig   `yaml:"remote"`
	Services ServicesConfig `yaml:"services"`

	Deployment DeploymentConfig `yaml:"deployment"`
}

func (c Config) FullImage() string {
	return c.Registry + "/" + c.Image + ":" + c.Tag
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	if cfg.Tag == "" {
		cfg.Tag = "latest"
	}
	if cfg.HealthPath == "" {
		cfg.HealthPath = "/health/"
	}
	if cfg.Proxy.Type == "" {
		cfg.Proxy.Type = "nginx"
	}
	if cfg.Proxy.Port == 0 {
		cfg.Proxy.Port = 8000
	}
	if cfg.AppPort == 0 {
		cfg.AppPort = 8000
	}
	if cfg.GreenPort == 0 {
		cfg.GreenPort = 8010
	}
	if cfg.BluePort == 0 {
		cfg.BluePort = 8011
	}
	if cfg.FaileoverTimeout == 0 {
		cfg.FaileoverTimeout,_ = time.ParseDuration("60s")
	}
	if cfg.ContainerDrainingTimeout == 0 {
		cfg.ContainerDrainingTimeout,_ = time.ParseDuration("30s")
	}
	if cfg.HealthCheckDelay == 0 {
		cfg.HealthCheckDelay,_ = time.ParseDuration("5s")
	}
	if cfg.NetworkName == "" {
		cfg.NetworkName = "app-network"
	}
	if cfg.Deployment.MonitorSeconds == 0 {
		cfg.Deployment.MonitorSeconds = 120
	}
	if cfg.Deployment.CheckIntervalSeconds == 0 {
		cfg.Deployment.CheckIntervalSeconds = 5
	}
	// by default enable auto-rollback
	if !cfg.Deployment.AutoRollback {
		// keep as-is if explicitly false in YAML, otherwise true
		cfg.Deployment.AutoRollback = true
	}
	return &cfg, nil
}