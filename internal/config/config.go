package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Queues []QueueConfig `yaml:"queues"`
}

type ServerConfig struct {
	GRPCPort int `yaml:"grpc_port"`
	HTTPPort int `yaml:"http_port"`
}

type QueueConfig struct {
	Name           string `yaml:"name"`
	Size           int    `yaml:"size"`
	MaxSubscribers int    `yaml:"max_subscribers"`
}

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