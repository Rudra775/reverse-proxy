package config

import (
	"encoding/json"
	"os"
)

type RouteConfig struct {
	PathPrefix string   `json:"path_prefix"`
	Backends   []string `json:"backends"`
	Strategy   string   `json:"strategy"` // "round_robin" or "least_conn"
}

type Config struct {
	ListenAddr       string        `json:"listen_addr"`
	Routes           []RouteConfig `json:"routes"`
	RequestTimeoutMs int           `json:"request_timeout_ms"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.RequestTimeoutMs <= 0 {
		cfg.RequestTimeoutMs = 5000
	}
	return &cfg, nil
}
