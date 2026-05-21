package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all driftlog runtime configuration.
type Config struct {
	StatePath    string        `yaml:"state_path"`
	OutputFormat string        `yaml:"output_format"`
	Region       string        `yaml:"region"`
	ResourceTypes []string     `yaml:"resource_types"`
	OnlyDrifted  bool          `yaml:"only_drifted"`
	Cache        CacheConfig   `yaml:"cache"`
}

// CacheConfig controls the optional local resource cache.
type CacheConfig struct {
	Enabled bool          `yaml:"enabled"`
	Path    string        `yaml:"path"`
	TTL     time.Duration `yaml:"ttl"`
}

// Default returns a Config populated with sensible defaults.
func Default() Config {
	return Config{
		StatePath:    "terraform.tfstate",
		OutputFormat: "text",
		Region:       "us-east-1",
		Cache: CacheConfig{
			Enabled: false,
			Path:    ".driftlog-cache.json",
			TTL:     5 * time.Minute,
		},
	}
}

// LoadFile reads a YAML config file and merges it over the defaults.
// If the file does not exist the defaults are returned without error.
func LoadFile(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("config: read file: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: parse yaml: %w", err)
	}
	valid := map[string]bool{"text": true, "json": true}
	if !valid[cfg.OutputFormat] {
		return cfg, fmt.Errorf("config: invalid output_format %q (must be text or json)", cfg.OutputFormat)
	}
	return cfg, nil
}
