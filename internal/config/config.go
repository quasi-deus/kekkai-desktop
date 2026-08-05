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
	AutoType   bool   `yaml:"auto_type" help:"Auto-type password instead of copying to clipboard" short:"a"`
	CopyPass   bool   `yaml:"copy_pass" help:"Copy password to clipboard" short:"c"`
	CopyUser   bool   `yaml:"copy_user" help:"Copy username to clipboard" short:"u"`
	CopyOTP    bool   `yaml:"copy_otp" help:"Copy TOTP 2FA code to clipboard" short:"o"`
	ViewNotes  bool   `yaml:"view_notes" help:"View entry notes" short:"n"`
	Delay      int    `yaml:"delay_ms" help:"Delay before auto-typing in milliseconds" default:"200"`
	ClearAfter int    `yaml:"clear_after_sec" help:"Clipboard auto-clear delay in seconds" default:"45"`
	Backend    string `yaml:"autotype_backend" enum:"auto,xdotool,wtype,ydotool" default:"auto" help:"Auto-type backend"`
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

// WriteDefaultConfig writes a sample config.yaml file if one does not already exist.
func WriteDefaultConfig(path string) (string, error) {
	expanded := ExpandPath(path)
	dir := filepath.Dir(expanded)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	defaultContent := `# rofi-pass-go Configuration File

# Path to the password store directory
store_dir: "~/.password-store"

# Auto-type password instead of copying to clipboard on Enter
auto_type: false

# Delay before auto-typing characters in milliseconds
delay_ms: 200

# Clipboard auto-clear timeout in seconds (0 to disable)
clear_after_sec: 45

# Auto-type backend: auto (autodetect Wayland/X11), xdotool, wtype, or ydotool
autotype_backend: "auto"
`

	if err := os.WriteFile(expanded, []byte(defaultContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write config file %s: %w", expanded, err)
	}

	return expanded, nil
}
