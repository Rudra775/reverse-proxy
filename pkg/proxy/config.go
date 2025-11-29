package proxy

import (
	"encoding/json"
	"os"
	"time"
)

type RouteConfig struct {
	PathPrefix string   `json:"path_prefix"`
	Backends   []string `json:"backends"`
	Strategy   string   `json:"strategy"`
}

type Config struct {
	ListenAddr       string        `json:"listen_addr"`
	Routes           []RouteConfig `json:"routes"`
	RequestTimeoutMs int           `json:"request_timeout_ms"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
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

func (c *Config) Timeout() time.Duration {
	return time.Duration(c.RequestTimeoutMs) * time.Millisecond
}
