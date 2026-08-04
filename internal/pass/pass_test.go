package pass_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quasi-deus/rofi-pass-go/internal/pass"
)

func TestParseEntryContent(t *testing.T) {
	content := `s3cr3tP@ssw0rd!
username: admin_john
url: https://github.com
otpauth://totp/GitHub:admin_john?secret=JBSWY3DPEHPK3PXP&issuer=GitHub
notes: Some extra notes line 1
Some extra notes line 2
`

	entry := pass.ParseEntryContent("github.com/admin_john", content)

	if entry.Password != "s3cr3tP@ssw0rd!" {
		t.Errorf("expected password 's3cr3tP@ssw0rd!', got '%s'", entry.Password)
	}
	if entry.Username != "admin_john" {
		t.Errorf("expected username 'admin_john', got '%s'", entry.Username)
	}
	if entry.URL != "https://github.com" {
		t.Errorf("expected url 'https://github.com', got '%s'", entry.URL)
	}
	if entry.ExtractOTPSecretKey() != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected OTP secret 'JBSWY3DPEHPK3PXP', got '%s'", entry.ExtractOTPSecretKey())
	}
	if entry.Notes == "" {
		t.Error("expected non-empty notes")
	}
}

func TestStoreListEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock password store files
	files := []string{
		"email/personal.gpg",
		"email/work.gpg",
		"github/account.gpg",
		".git/config",
		".gpg-id",
	}

	for _, f := range files {
		fullPath := filepath.Join(tmpDir, f)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte("mock"), 0644)
	}

	store := pass.NewStore(tmpDir)
	entries, err := store.ListEntries()

	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}

	expected := []string{
		"email/personal",
		"email/work",
		"github/account",
	}

	if len(entries) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(entries))
	}

	for i, e := range expected {
		if entries[i] != e {
			t.Errorf("entry %d: expected %s, got %s", i, e, entries[i])
		}
	}
}
