package rofi_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quasi-deus/rofi-pass-go/internal/config"
	"github.com/quasi-deus/rofi-pass-go/internal/pass"
	"github.com/quasi-deus/rofi-pass-go/internal/rofi"
)

func TestHandleScriptModeInitialLoad(t *testing.T) {
	tmpDir := t.TempDir()

	entryPath := filepath.Join(tmpDir, "github.com/alice.gpg")
	os.MkdirAll(filepath.Dir(entryPath), 0755)
	os.WriteFile(entryPath, []byte("mock"), 0644)

	cfg := &config.Config{
		StoreDir:   tmpDir,
		ClearAfter: 0,
	}

	os.Setenv("ROFI_RETV", "0")
	defer os.Unsetenv("ROFI_RETV")

	err := rofi.HandleScriptMode(cfg, "")
	if err != nil {
		t.Fatalf("HandleScriptMode failed: %v", err)
	}
}

func TestCopyPasswordNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := pass.NewStore(tmpDir)

	cfg := &config.Config{
		StoreDir:   tmpDir,
		ClearAfter: 0,
	}

	err := rofi.CopyPassword("nonexistent", cfg, store)
	if err == nil {
		t.Error("expected error for nonexistent entry")
	}
}
