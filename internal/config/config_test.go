package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quasi-deus/kekkai-desktop/internal/config"
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

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
store_dir: "~/.password-store"
clear_after_sec: 45
`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := config.LoadYAML(configPath)
	if err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}

	if cfg.StoreDir != "~/.password-store" {
		t.Errorf("expected store_dir ~/.password-store, got %s", cfg.StoreDir)
	}
	if cfg.ClearAfter != 45 {
		t.Errorf("expected clear_after_sec 45, got %d", cfg.ClearAfter)
	}
}
