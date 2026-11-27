// Package config loads gateway settings from environment variables.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds runtime settings for the gateway.
type Config struct {
	ListenAddr      string
	UpstreamURL     string
	DisplayCurrency string
	ShutdownTimeout time.Duration
}

// FromEnv reads configuration from TOKENGATE_* variables, applying defaults.
func FromEnv() (Config, error) {
	c := Config{
		ListenAddr:      get("TOKENGATE_LISTEN_ADDR", ":8080"),
		UpstreamURL:     get("TOKENGATE_UPSTREAM_URL", "https://api.anthropic.com"),
		DisplayCurrency: get("TOKENGATE_DISPLAY_CURRENCY", "INR"),
		ShutdownTimeout: 10 * time.Second,
	}
	if v := os.Getenv("TOKENGATE_SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("TOKENGATE_SHUTDOWN_TIMEOUT: %w", err)
		}
		c.ShutdownTimeout = d
	}
	return c, nil
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
