package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/quasi-deus/rofi-pass-go/internal/autotype"
	"github.com/quasi-deus/rofi-pass-go/internal/clipboard"
	"github.com/quasi-deus/rofi-pass-go/internal/config"
	"github.com/quasi-deus/rofi-pass-go/internal/pass"
	"github.com/quasi-deus/rofi-pass-go/internal/rofi"
	"github.com/quasi-deus/rofi-pass-go/internal/totp"
)

type CLI struct {
	StoreDir      string `help:"Path to password store directory" default:"~/.password-store" env:"PASSWORD_STORE_DIR"`
	ConfigFile    string `help:"Path to YAML configuration file" short:"C" type:"path" default:"~/.config/rofi-pass-go/config.yaml"`
	AutoType      bool   `help:"Auto-type password instead of copying to clipboard" short:"a"`
	CopyPass      bool   `help:"Copy password to clipboard" short:"c"`
	CopyUser      bool   `help:"Copy username to clipboard" short:"u"`
	CopyOTP       bool   `help:"Copy TOTP 2FA code to clipboard" short:"o"`
	ViewNotes     bool   `help:"View entry notes" short:"n"`
	Delay         int    `help:"Delay before auto-typing in milliseconds" default:"200"`
	ClearAfter    int    `help:"Clipboard auto-clear delay in seconds" default:"45"`
	Backend       string `enum:"auto,xdotool,wtype,ydotool" default:"auto" help:"Auto-type backend"`
	EnableSubmenu bool   `help:"Enable interactive sub-menu"`
	DefaultAction string `help:"Default action (copy-pass, autotype, copy-user, copy-otp, submenu)" default:"copy-pass"`

	// Subcommands
	Rofi       RofiCmd       `cmd:"" default:"1" help:"Run in Rofi script mode (default)"`
	List       ListCmd       `cmd:"" help:"List all password entries"`
	Show       ShowCmd       `cmd:"" help:"Show/decrypt a specific password entry"`
	Copy       CopyCmd       `cmd:"" help:"Copy password, username, or OTP for an entry"`
	Type       TypeCmd       `cmd:"" help:"Auto-type password for an entry"`
	InitConfig InitConfigCmd `cmd:"" help:"Generate default YAML config file"`
}

type RofiCmd struct{}
type ListCmd struct{}
type ShowCmd struct {
	Entry string `arg:"" help:"Pass entry path (e.g. github.com/user)"`
}
type CopyCmd struct {
	Entry string `arg:"" help:"Pass entry path"`
	Field string `enum:"pass,user,otp" default:"pass" help:"Field to copy (pass, user, otp)"`
}
type TypeCmd struct {
	Entry string `arg:"" help:"Pass entry path"`
}
type InitConfigCmd struct {
	Path string `help:"Path where to generate config.yaml" default:"~/.config/rofi-pass-go/config.yaml"`
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

	knownCommands := map[string]bool{
		"rofi":        true,
		"list":        true,
		"show":        true,
		"copy":        true,
		"type":        true,
		"init-config": true,
		"help":        true,
	}

	if !knownCommands[firstArg] {
		// Store selection and strip it from args so Kong parses default RofiCmd cleanly
		extraRofiSelection = firstArg
		if os.Getenv("ROFI_INFO") == "" {
			os.Setenv("ROFI_INFO", firstArg)
		}
		newArgs := append([]string{args[0]}, args[2:]...)
		return newArgs
	}

	return args
}

func main() {
	var cli CLI

	processedArgs := preprocessArgs(os.Args)

	parser, err := kong.New(&cli,
		kong.Name("rofi-pass-go"),
		kong.Description("Fast, Go-native Rofi interface for pass (password-store) with Wayland/X11 and YAML support."),
		kong.UsageOnError(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CLI parser: %v\n", err)
		os.Exit(1)
	}

	ctx, err := parser.Parse(processedArgs[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	// Build app Config by loading YAML file if present and merging CLI values
	appConfig := buildAppConfig(&cli)

	err = ctx.Run(appConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func buildAppConfig(cli *CLI) *config.Config {
	cfg := &config.Config{
		StoreDir:      cli.StoreDir,
		AutoType:      cli.AutoType,
		CopyPass:      cli.CopyPass,
		CopyUser:      cli.CopyUser,
		CopyOTP:       cli.CopyOTP,
		ViewNotes:     cli.ViewNotes,
		Delay:         cli.Delay,
		ClearAfter:    cli.ClearAfter,
		Backend:       cli.Backend,
		EnableSubmenu: cli.EnableSubmenu,
		DefaultAction: cli.DefaultAction,
		ConfigFile:    cli.ConfigFile,
	}

	// Read YAML config file if it exists
	yamlCfg, err := config.LoadYAML(cli.ConfigFile)
	if err == nil && yamlCfg != nil {
		if yamlCfg.StoreDir != "" && cli.StoreDir == "~/.password-store" {
			cfg.StoreDir = yamlCfg.StoreDir
		}
		if yamlCfg.Delay > 0 && cli.Delay == 200 {
			cfg.Delay = yamlCfg.Delay
		}
		if yamlCfg.ClearAfter >= 0 && cli.ClearAfter == 45 {
			cfg.ClearAfter = yamlCfg.ClearAfter
		}
		if yamlCfg.Backend != "" && cli.Backend == "auto" {
			cfg.Backend = yamlCfg.Backend
		}
		if yamlCfg.DefaultAction != "" && cli.DefaultAction == "copy-pass" {
			cfg.DefaultAction = yamlCfg.DefaultAction
		}
		if yamlCfg.EnableSubmenu {
			cfg.EnableSubmenu = true
		}
		cfg.Hotkeys = yamlCfg.Hotkeys
	}

	return cfg
}

func (r *RofiCmd) Run(cfg *config.Config) error {
	return rofi.HandleScriptMode(cfg, extraRofiSelection)
}

func (l *ListCmd) Run(cfg *config.Config) error {
	store := pass.NewStore(cfg.StoreDir)
	entries, err := store.ListEntries()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fmt.Println(entry)
	}
	return nil
}

func (s *ShowCmd) Run(cfg *config.Config) error {
	store := pass.NewStore(cfg.StoreDir)
	entry, err := store.GetEntry(s.Entry)
	if err != nil {
		return err
	}

	fmt.Printf("Entry: %s\n", entry.Name)
	fmt.Printf("Password: %s\n", entry.Password)
	if entry.Username != "" {
		fmt.Printf("Username: %s\n", entry.Username)
	}
	if entry.URL != "" {
		fmt.Printf("URL: %s\n", entry.URL)
	}
	if secret := entry.ExtractOTPSecretKey(); secret != "" {
		code, _ := totp.GenerateCode(secret)
		fmt.Printf("OTP Code: %s (Secret: %s)\n", code, secret)
	}
	if entry.Notes != "" {
		fmt.Printf("Notes:\n%s\n", entry.Notes)
	}
	return nil
}

func (c *CopyCmd) Run(cfg *config.Config) error {
	store := pass.NewStore(cfg.StoreDir)
	entry, err := store.GetEntry(c.Entry)
	if err != nil {
		return err
	}

	var targetText string
	switch c.Field {
	case "pass":
		targetText = entry.Password
	case "user":
		targetText = entry.Username
	case "otp":
		secret := entry.ExtractOTPSecretKey()
		if secret == "" {
			return fmt.Errorf("no OTP secret found in entry %s", c.Entry)
		}
		targetText, err = totp.GenerateCode(secret)
		if err != nil {
			return err
		}
	}

	if targetText == "" {
		return fmt.Errorf("field %s is empty for entry %s", c.Field, c.Entry)
	}

	return clipboard.CopyAndClear(targetText, cfg.ClearAfter)
}

func (t *TypeCmd) Run(cfg *config.Config) error {
	store := pass.NewStore(cfg.StoreDir)
	entry, err := store.GetEntry(t.Entry)
	if err != nil {
		return err
	}

	return autotype.TypeText(entry.Password, cfg.Delay, autotype.Backend(cfg.Backend))
}

func (i *InitConfigCmd) Run() error {
	createdPath, err := config.WriteDefaultConfig(i.Path)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully generated configuration file at %s\n", createdPath)
	return nil
}
