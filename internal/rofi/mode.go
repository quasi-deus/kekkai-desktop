package rofi

import (
	"fmt"
	"os"
	"os/exec"

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
		selectedEntry := info
		if selectedEntry == "" {
			selectedEntry = selectedArg
		}

		if selectedEntry == "" {
			return nil
		}

		action := determineAction(retv, cfg)

		// Launch action in a detached background process so Rofi exits immediately
		// and releases the Wayland keyboard focus grab for GPG pinentry.
		return ExecuteActionDetached(action, selectedEntry)

	default:
		return nil
	}
}

func determineAction(retv string, cfg *config.Config) Action {
	switch retv {
	case "1": // Enter
		if cfg.AutoType {
			return ActionAutoType
		}
		return ActionCopyPass
	case "10": // Custom 1
		return ActionCopyPass
	case "11": // Custom 2
		return ActionCopyUser
	case "12": // Custom 3
		return ActionCopyOTP
	case "13": // Custom 4
		return ActionAutoType
	case "14": // Custom 5
		return ActionViewNotes
	default:
		return ActionCopyPass
	}
}

// ExecuteActionDetached launches the target action in a background process so Rofi releases Wayland focus.
func ExecuteActionDetached(action Action, entryName string) error {
	selfBinary, err := os.Executable()
	if err != nil {
		selfBinary = "rofi-pass-go"
	}

	var args []string
	switch action {
	case ActionCopyPass:
		args = []string{"copy", entryName, "--field", "pass"}
	case ActionCopyUser:
		args = []string{"copy", entryName, "--field", "user"}
	case ActionCopyOTP:
		args = []string{"copy", entryName, "--field", "otp"}
	case ActionAutoType:
		args = []string{"type", entryName}
	case ActionViewNotes:
		args = []string{"show", entryName}
	default:
		args = []string{"copy", entryName, "--field", "pass"}
	}

	cmd := exec.Command(selfBinary, args...)
	cmd.Env = os.Environ()
	return cmd.Start()
}

// ExecuteAction decrypts the entry and performs the target action directly.
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
			return nil
		}
		return clipboard.CopyAndClear(entry.Username, cfg.ClearAfter)

	case ActionCopyOTP:
		secret := entry.ExtractOTPSecretKey()
		if secret == "" {
			return nil
		}
		code, err := totp.GenerateCode(secret)
		if err != nil {
			return err
		}
		return clipboard.CopyAndClear(code, cfg.ClearAfter)

	case ActionAutoType:
		return autotype.TypeText(entry.Password, cfg.Delay, autotype.Backend(cfg.Backend))

	case ActionViewNotes:
		if entry.Notes != "" {
			fmt.Println(entry.Notes)
		}
		return nil

	default:
		return clipboard.CopyAndClear(entry.Password, cfg.ClearAfter)
	}
}
