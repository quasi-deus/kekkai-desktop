package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quasi-deus/rofi-pass-go/internal/config"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("User home dir not available")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/.password-store", filepath.Join(home, ".password-store")},
		{"/abs/path", "/abs/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := config.ExpandPath(tt.input)
		if got != tt.expected {
			t.Errorf("ExpandPath(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestWriteAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sub", "config.yaml")

	// Test writing default config
	createdPath, err := config.WriteDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("WriteDefaultConfig failed: %v", err)
	}
	if createdPath != configPath {
		t.Errorf("expected path %s, got %s", configPath, createdPath)
	}

	// Test loading written YAML config
	cfg, err := config.LoadYAML(configPath)
	if err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}

	if cfg.StoreDir != "~/.password-store" {
		t.Errorf("expected store_dir ~/.password-store, got %s", cfg.StoreDir)
	}
	if cfg.Delay != 200 {
		t.Errorf("expected delay_ms 200, got %d", cfg.Delay)
	}
	if cfg.ClearAfter != 45 {
		t.Errorf("expected clear_after_sec 45, got %d", cfg.ClearAfter)
	}
	if cfg.Backend != "auto" {
		t.Errorf("expected backend auto, got %s", cfg.Backend)
	}
}
