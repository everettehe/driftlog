package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/user/driftlog/internal/notify"
)

// OutputFormat represents the supported output formats.
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// Config holds all driftlog runtime configuration.
type Config struct {
	StatePath    string            `yaml:"state_path"`
	Region       string            `yaml:"region"`
	OutputFormat OutputFormat      `yaml:"output_format"`
	CacheTTL     string            `yaml:"cache_ttl"`
	Schedule     string            `yaml:"schedule"`
	Filter       FilterConfig      `yaml:"filter"`
	Notify       notify.Config     `yaml:"notify"`
}

// FilterConfig controls which resources are included in output.
type FilterConfig struct {
	OnlyDrift bool     `yaml:"only_drift"`
	Types     []string `yaml:"types"`
}

// Default returns a Config populated with sensible defaults.
func Default() Config {
	return Config{
		StatePath:    "terraform.tfstate",
		Region:       "us-east-1",
		OutputFormat: FormatText,
		CacheTTL:     "5m",
		Schedule:     "",
	}
}

// LoadFile reads a YAML config file and merges it over the defaults.
// If the file does not exist, defaults are returned without error.
func LoadFile(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("config: read file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: parse yaml: %w", err)
	}

	if cfg.OutputFormat != FormatText && cfg.OutputFormat != FormatJSON {
		return cfg, fmt.Errorf("config: invalid output_format %q (must be 'text' or 'json')", cfg.OutputFormat)
	}

	return cfg, nil
}
