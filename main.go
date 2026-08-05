package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/quasi-deus/rofi-pass-go/internal/config"
	"github.com/quasi-deus/rofi-pass-go/internal/pass"
	"github.com/quasi-deus/rofi-pass-go/internal/rofi"
)

type CLI struct {
	StoreDir   string `help:"Path to password store directory" default:"~/.password-store" env:"PASSWORD_STORE_DIR"`
	ConfigFile string `help:"Path to YAML configuration file" short:"C" type:"path" default:"~/.config/rofi-pass-go/config.yaml"`
	ClearAfter int    `help:"Clipboard auto-clear delay in seconds" default:"45"`
	CopyAction string `help:"Internal flag to copy password for entry" hidden:""`

	Selected string `arg:"" optional:"" help:"Selected entry path from Rofi"`
}

func main() {
	var cli CLI

	kong.Parse(&cli,
		kong.Name("rofi-pass-go"),
		kong.Description("Fast, Go-native Rofi interface for pass."),
		kong.UsageOnError(),
	)

	appConfig := buildAppConfig(&cli)

	// If --copy-action is set (detached background process execution)
	if cli.CopyAction != "" {
		store := pass.NewStore(appConfig.StoreDir)
		err := rofi.CopyPassword(cli.CopyAction, appConfig, store)
		if err != nil {
			os.Exit(1)
		}
		return
	}

	// Default: Rofi script mode
	err := rofi.HandleScriptMode(appConfig, cli.Selected)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func buildAppConfig(cli *CLI) *config.Config {
	cfg := &config.Config{
		StoreDir:   cli.StoreDir,
		ClearAfter: cli.ClearAfter,
		ConfigFile: cli.ConfigFile,
	}

	yamlCfg, err := config.LoadYAML(cli.ConfigFile)
	if err == nil && yamlCfg != nil {
		if yamlCfg.StoreDir != "" && cli.StoreDir == "~/.password-store" {
			cfg.StoreDir = yamlCfg.StoreDir
		}
		if yamlCfg.ClearAfter >= 0 && cli.ClearAfter == 45 {
			cfg.ClearAfter = yamlCfg.ClearAfter
		}
	}

	return cfg
}
