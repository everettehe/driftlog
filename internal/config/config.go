package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents the report output format.
type OutputFormat string

const (
	OutputText OutputFormat = "text"
	OutputJSON OutputFormat = "json"
)

// Config holds all driftlog runtime configuration.
type Config struct {
	StateFile    string       `yaml:"state_file"`
	OutputFormat OutputFormat `yaml:"output_format"`
	AWS          AWSConfig    `yaml:"aws"`
	Filter       FilterConfig `yaml:"filter"`
}

// AWSConfig holds AWS-specific settings.
type AWSConfig struct {
	Region  string   `yaml:"region"`
	Profile string   `yaml:"profile"`
	Services []string `yaml:"services"`
}

// FilterConfig maps directly to filter.Options fields for YAML loading.
type FilterConfig struct {
	OnlyDrifted   bool     `yaml:"only_drifted"`
	ResourceTypes []string `yaml:"resource_types"`
	ExcludeIDs    []string `yaml:"exclude_ids"`
}

// Default returns a Config populated with sensible defaults.
func Default() Config {
	return Config{
		StateFile:    "terraform.tfstate",
		OutputFormat: OutputText,
		AWS: AWSConfig{
			Region:   "us-east-1",
			Services: []string{"ec2", "s3"},
		},
	}
}

// LoadFile reads a YAML config file and merges it over the defaults.
// If the file does not exist the defaults are returned without error.
func LoadFile(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	if cfg.OutputFormat != OutputText && cfg.OutputFormat != OutputJSON {
		return cfg, errors.New("invalid output_format: must be \"text\" or \"json\"")
	}

	return cfg, nil
}
