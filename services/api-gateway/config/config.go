package config

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	instance *Config
	once     sync.Once
)

// Service name constants — used as keys in Targets map
// and as identifiers throughout the codebase
const (
	AdService    = "ad-service"
	OrderService = "order-service"
)

// Instance returns the loaded, validated config singleton.
// Safe for concurrent use. Panics on misconfiguration —
// a service that cannot read its config must not start.
func Instance() *Config {
	once.Do(func() {
		cfg, err := load()
		if err != nil {
			panic(fmt.Sprintf("config: failed to load: %s", err))
		}
		instance = cfg
	})
	return instance
}

func load() (*Config, error) {
	cfg := &Config{}

	// CONFIG_PATH env var allows Docker/k8s to inject config location
	// without rebuilding the image
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml" // sensible default for local dev
	}

	// Layer 1: yaml file provides base config with defaults
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		return nil, fmt.Errorf("read yaml: %w", err)
	}
	// Layer 2: env vars override yaml — enables 12-factor config
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("read env: %w", err)
	}
	// Layer 3: validate — catch zero-values and missing required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.GrpcServer.Port == 0 {
		return errors.New("grpc_server.port is required")
	}
	if c.HttpServer.Port == 0 {
		return errors.New("http_server.port is required")
	}
	if c.HttpServer.AdminPort == 0 {
		return errors.New("http_server.admin_port is required")
	}
	if c.Graceful.Timeout == 0 {
		return errors.New("graceful.timeout is required")
	}
	if len(c.Targets) == 0 {
		return errors.New("no service targets configured")
	}
	return nil
}
