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
	ListenAddr             string        `json:"listen_addr"`
	Routes                 []RouteConfig `json:"routes"`
	RequestTimeoutMs       int           `json:"request_timeout_ms"`
	MaxRetries             int           `json:"max_retries"`
	CacheSizeMB            int           `json:"cache_size_mb"`
	HealthCheckIntervalSec int           `json:"health_check_interval_sec"`
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
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.CacheSizeMB < 0 {
		cfg.CacheSizeMB = 0
	}
	if cfg.HealthCheckIntervalSec < 0 {
		cfg.HealthCheckIntervalSec = 0
	}

	return &cfg, nil
}

func (c *Config) Timeout() time.Duration {
	return time.Duration(c.RequestTimeoutMs) * time.Millisecond
}

func (c *Config) CacheBytes() int {
	if c.CacheSizeMB == 0 {
		return 0
	}
	return c.CacheSizeMB * 1024 * 1024
}

func (c *Config) HealthCheckInterval() time.Duration {
	if c.HealthCheckIntervalSec <= 0 {
		return 0
	}
	return time.Duration(c.HealthCheckIntervalSec) * time.Second
}
