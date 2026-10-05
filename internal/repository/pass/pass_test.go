package pass_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quasi-deus/kekkai/desktop/internal/repository/pass"
)

func TestParsePassword(t *testing.T) {
	content := `s3cr3tP@ssw0rd!
username: admin_john
url: https://github.com
`

	pwd := pass.ParsePassword(content)
	if pwd != "s3cr3tP@ssw0rd!" {
		t.Errorf("expected password 's3cr3tP@ssw0rd!', got '%s'", pwd)
	}
}

func TestStoreListEntries(t *testing.T) {
	tmpDir := t.TempDir()

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
