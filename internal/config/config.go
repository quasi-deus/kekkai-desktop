package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration schema.
type Config struct {
	StoreDir   string `yaml:"store_dir" help:"Path to password store directory" default:"~/.password-store" env:"PASSWORD_STORE_DIR"`
	ClearAfter int    `yaml:"clear_after_sec" help:"Clipboard auto-clear delay in seconds" default:"45"`
	ConfigFile string `yaml:"-" help:"Path to YAML configuration file" type:"path" short:"C"`
}

// ExpandPath replaces leading ~ with the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// LoadYAML reads and parses a YAML configuration file.
func LoadYAML(path string) (*Config, error) {
	expanded := ExpandPath(path)
	data, err := os.ReadFile(expanded)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config at %s: %w", path, err)
	}

	return &cfg, nil
}
