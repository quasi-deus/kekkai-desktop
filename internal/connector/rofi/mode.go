package rofi

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/quasi-deus/kekkai/desktop/internal/clipboard"
	"github.com/quasi-deus/kekkai/desktop/internal/config"
	"github.com/quasi-deus/kekkai/desktop/internal/repository"
	"github.com/quasi-deus/kekkai/desktop/internal/repository/pass"
)

// HandleScriptMode orchestrates Rofi script mode protocol.
func HandleScriptMode(cfg *config.Config, selectedArg string) error {
	retv := os.Getenv("ROFI_RETV")
	info := os.Getenv("ROFI_INFO")

	store := pass.NewStore(cfg.StoreDir)

	switch retv {
	case "0", "":
		// Initial Load: Print entry list for Rofi
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

		// Detach copy action into background process so Rofi exits immediately
		// and releases Wayland keyboard focus for GPG pinentry.
		return CopyPasswordDetached(selectedEntry)

	default:
		return nil
	}
}

// CopyPasswordDetached launches the copy action in a background process.
func CopyPasswordDetached(entryName string) error {
	selfBinary, err := os.Executable()
	if err != nil {
		selfBinary = "kekkai-rofi"
	}

	cmd := exec.Command(selfBinary, "--copy-action", entryName)
	cmd.Env = os.Environ()
	return cmd.Start()
}

// CopyPassword decrypts the entry and copies password to clipboard.
func CopyPassword(entryName string, cfg *config.Config, store repository.Repository) error {
	pwd, err := store.GetPassword(entryName)
	if err != nil {
		fmt.Printf("Failed to decrypt entry %s: %v\n", entryName, err)
		return err
	}

	return clipboard.CopyAndClear(pwd, cfg.ClearAfter)
}
