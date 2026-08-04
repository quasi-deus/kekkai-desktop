package pass

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/quasi-deus/rofi-pass-go/internal/config"
)

// Store interacts with the local password-store directory (~/.password-store).
type Store struct {
	Dir string
}

// NewStore creates a new Store instance given a store directory path.
func NewStore(dir string) *Store {
	expanded := config.ExpandPath(dir)
	return &Store{Dir: expanded}
}

// ListEntries returns a sorted slice of relative entry paths (without .gpg extension).
func (s *Store) ListEntries() ([]string, error) {
	if _, err := os.Stat(s.Dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("password store directory does not exist: %s", s.Dir)
	}

	var entries []string

	err := filepath.WalkDir(s.Dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories like .git, .gpg-id, .extensions
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") && path != s.Dir {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		// Ignore non-.gpg files or hidden files
		if strings.HasPrefix(d.Name(), ".") || !strings.HasSuffix(d.Name(), ".gpg") {
			return nil
		}

		rel, err := filepath.Rel(s.Dir, path)
		if err != nil {
			return err
		}

		// Trim .gpg suffix
		entryName := strings.TrimSuffix(rel, ".gpg")
		entries = append(entries, entryName)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk password store directory: %w", err)
	}

	sort.Strings(entries)
	return entries, nil
}

// GetEntry decrypts and parses a specific pass entry.
func (s *Store) GetEntry(name string) (*Entry, error) {
	// Sanitize entry name
	cleanName := strings.TrimSuffix(name, ".gpg")
	filePath := filepath.Join(s.Dir, cleanName+".gpg")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("entry does not exist: %s", cleanName)
	}

	// Decrypt GPG file using gpg CLI
	cmd := exec.Command("gpg", "--decrypt", "--batch", "--quiet", filePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to decrypt %s: %s (%w)", cleanName, stderr.String(), err)
	}

	return ParseEntryContent(cleanName, stdout.String()), nil
}
