package rofi

import (
	"fmt"
	"os"
	"strings"

	"github.com/quasi-deus/rofi-pass-go/internal/autotype"
	"github.com/quasi-deus/rofi-pass-go/internal/clipboard"
	"github.com/quasi-deus/rofi-pass-go/internal/config"
	"github.com/quasi-deus/rofi-pass-go/internal/pass"
	"github.com/quasi-deus/rofi-pass-go/internal/totp"
)

// Action defines what action to perform on a pass entry.
type Action string

const (
	ActionCopyPass  Action = "copy-pass"
	ActionCopyUser  Action = "copy-user"
	ActionCopyOTP   Action = "copy-otp"
	ActionAutoType  Action = "autotype"
	ActionViewNotes Action = "view-notes"
	ActionSubmenu   Action = "submenu"
)

// HandleScriptMode orchestrates Rofi script mode protocol.
func HandleScriptMode(cfg *config.Config, selectedArg string) error {
	retv := os.Getenv("ROFI_RETV")
	info := os.Getenv("ROFI_INFO")

	store := pass.NewStore(cfg.StoreDir)

	switch retv {
	case "0", "":
		// Initial Load: Print plain entry lines to stdout for Rofi
		entries, err := store.ListEntries()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}

		for _, entry := range entries {
			fmt.Println(entry)
		}
		return nil

	case "1", "10", "11", "12", "13", "14":
		// User selected an entry or triggered a keybinding
		selectedEntry := ""
		if info != "" {
			selectedEntry = info
		} else if selectedArg != "" {
			selectedEntry = selectedArg
		}

		if selectedEntry == "" {
			return nil
		}

		// Check if we are inside a submenu selection
		if strings.HasPrefix(selectedEntry, "SUBMENU::") {
			return handleSubmenuSelection(selectedEntry, cfg, store)
		}

		action := determineAction(retv, cfg)

		if action == ActionSubmenu || cfg.EnableSubmenu {
			printSubmenu(selectedEntry)
			return nil
		}

		return ExecuteAction(action, selectedEntry, cfg, store)

	default:
		return nil
	}
}

func determineAction(retv string, cfg *config.Config) Action {
	switch retv {
	case "1": // Enter
		return Action(cfg.DefaultAction)
	case "10": // Alt+c / Custom 1
		return ActionCopyPass
	case "11": // Alt+u / Custom 2
		return ActionCopyUser
	case "12": // Alt+o / Custom 3
		return ActionCopyOTP
	case "13": // Alt+a / Custom 4
		return ActionAutoType
	case "14": // Alt+n / Custom 5
		return ActionViewNotes
	default:
		return ActionCopyPass
	}
}

func printSubmenu(entryName string) {
	fmt.Printf("[1] Copy Password\x00info\x1fSUBMENU::%s::copy-pass\n", entryName)
	fmt.Printf("[2] Copy Username\x00info\x1fSUBMENU::%s::copy-user\n", entryName)
	fmt.Printf("[3] Copy TOTP 2FA\x00info\x1fSUBMENU::%s::copy-otp\n", entryName)
	fmt.Printf("[4] Auto-type Password\x00info\x1fSUBMENU::%s::autotype\n", entryName)
	fmt.Printf("[5] View Notes\x00info\x1fSUBMENU::%s::view-notes\n", entryName)
}

func handleSubmenuSelection(selectedInfo string, cfg *config.Config, store *pass.Store) error {
	parts := strings.Split(selectedInfo, "::")
	if len(parts) < 3 {
		return fmt.Errorf("invalid submenu info format: %s", selectedInfo)
	}

	entryName := parts[1]
	action := Action(parts[2])

	return ExecuteAction(action, entryName, cfg, store)
}

// ExecuteAction decrypts the entry and performs the target action.
func ExecuteAction(action Action, entryName string, cfg *config.Config, store *pass.Store) error {
	entry, err := store.GetEntry(entryName)
	if err != nil {
		fmt.Printf("Failed to decrypt entry %s: %v\n", entryName, err)
		return err
	}

	switch action {
	case ActionCopyPass:
		return clipboard.CopyAndClear(entry.Password, cfg.ClearAfter)

	case ActionCopyUser:
		if entry.Username == "" {
			fmt.Println("No username found in entry")
			return nil
		}
		return clipboard.CopyAndClear(entry.Username, cfg.ClearAfter)

	case ActionCopyOTP:
		secret := entry.ExtractOTPSecretKey()
		if secret == "" {
			fmt.Println("No OTP secret found in entry")
			return nil
		}
		code, err := totp.GenerateCode(secret)
		if err != nil {
			fmt.Printf("Failed to generate TOTP code: %v\n", err)
			return err
		}
		return clipboard.CopyAndClear(code, cfg.ClearAfter)

	case ActionAutoType:
		return autotype.TypeText(entry.Password, cfg.Delay, autotype.Backend(cfg.Backend))

	case ActionViewNotes:
		if entry.Notes == "" {
			fmt.Println("No notes found in entry")
			return nil
		}
		fmt.Println(entry.Notes)
		return nil

	default:
		return clipboard.CopyAndClear(entry.Password, cfg.ClearAfter)
	}
}
