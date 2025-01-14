package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the main application config.
type Config struct {
	Server ServerConfig  `yaml:"server"`
	Queues []QueueConfig `yaml:"queues"`
}

// ServerConfig is the server config.
type ServerConfig struct {
	GRPCPort int `yaml:"grpc_port"`
	HTTPPort int `yaml:"http_port"`
}

// QueueConfig is the queue config.
type QueueConfig struct {
	Name           string `yaml:"name"`
	Size           int    `yaml:"size"`
	MaxSubscribers int    `yaml:"max_subscribers"`
}

// Load loads the config from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// GRPCAddress returns the address for the GRPC server.
func (c *Config) GRPCAddress() string {
	return fmt.Sprintf("0.0.0.0:%d", c.Server.GRPCPort)
}

// HTTPAddress returns the address for the HTTP server.
func (c *Config) HTTPAddress() string {
	return fmt.Sprintf("0.0.0.0:%d", c.Server.HTTPPort)
}
