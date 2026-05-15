package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all runtime configuration for driftlog.
type Config struct {
	StateFile  string   `yaml:"state_file"`
	Region     string   `yaml:"region"`
	Profile    string   `yaml:"profile"`
	OutputFmt  string   `yaml:"output_format"`
	Resources  []string `yaml:"resources"`
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		StateFile: "terraform.tfstate",
		Region:    "us-east-1",
		OutputFmt: "text",
		Resources: []string{"aws_instance", "aws_s3_bucket"},
	}
}

// LoadFile reads a YAML config file from the given path and merges it
// over the default configuration.
func LoadFile(path string) (*Config, error) {
	cfg := Default()

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that required fields contain acceptable values.
func (c *Config) Validate() error {
	if c.StateFile == "" {
		return errors.New("config: state_file must not be empty")
	}
	if c.Region == "" {
		return errors.New("config: region must not be empty")
	}
	switch c.OutputFmt {
	case "text", "json":
		// valid
	default:
		return errors.New("config: output_format must be \"text\" or \"json\"")
	}
	return nil
}
