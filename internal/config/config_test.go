package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/driftlog/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "driftlog.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return p
}

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.StateFile != "terraform.tfstate" {
		t.Errorf("expected default state_file, got %q", cfg.StateFile)
	}
	if cfg.OutputFmt != "text" {
		t.Errorf("expected default output_format \"text\", got %q", cfg.OutputFmt)
	}
}

func TestLoadFile_MissingFile_ReturnsDefaults(t *testing.T) {
	cfg, err := config.LoadFile("/nonexistent/path/driftlog.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Region != "us-east-1" {
		t.Errorf("expected default region, got %q", cfg.Region)
	}
}

func TestLoadFile_ValidYAML(t *testing.T) {
	p := writeTempConfig(t, "state_file: prod.tfstate\nregion: eu-west-1\noutput_format: json\n")
	cfg, err := config.LoadFile(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.StateFile != "prod.tfstate" {
		t.Errorf("state_file: got %q", cfg.StateFile)
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("region: got %q", cfg.Region)
	}
	if cfg.OutputFmt != "json" {
		t.Errorf("output_format: got %q", cfg.OutputFmt)
	}
}

func TestLoadFile_InvalidOutputFormat(t *testing.T) {
	p := writeTempConfig(t, "output_format: xml\n")
	_, err := config.LoadFile(p)
	if err == nil {
		t.Fatal("expected validation error for invalid output_format")
	}
}

func TestValidate_EmptyStateFile(t *testing.T) {
	cfg := config.Default()
	cfg.StateFile = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty state_file")
	}
}
