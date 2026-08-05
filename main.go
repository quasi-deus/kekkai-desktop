package main

import (
	"fmt"
	"os"
	"strings"

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
}

var extraRofiSelection string

func preprocessArgs(args []string) []string {
	if len(args) <= 1 {
		return args
	}

	firstArg := args[1]
	if strings.HasPrefix(firstArg, "-") {
		return args
	}

	// Positional argument passed by Rofi upon selection
	extraRofiSelection = firstArg
	if os.Getenv("ROFI_INFO") == "" {
		os.Setenv("ROFI_INFO", firstArg)
	}
	return append([]string{args[0]}, args[2:]...)
}

func main() {
	var cli CLI

	processedArgs := preprocessArgs(os.Args)

	parser, err := kong.New(&cli,
		kong.Name("rofi-pass-go"),
		kong.Description("Fast, Go-native Rofi interface for pass."),
		kong.UsageOnError(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CLI parser: %v\n", err)
		os.Exit(1)
	}

	_, err = parser.Parse(processedArgs[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

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
	err = rofi.HandleScriptMode(appConfig, extraRofiSelection)
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
