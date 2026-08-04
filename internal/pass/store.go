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
	resolved, err := filepath.EvalSymlinks(expanded)
	if err == nil {
		expanded = resolved
	}
	return &Store{Dir: expanded}
}

// ListEntries returns a sorted slice of relative entry paths (without .gpg extension).
func (s *Store) ListEntries() ([]string, error) {
	realDir, err := filepath.EvalSymlinks(s.Dir)
	if err != nil {
		realDir = s.Dir
	}

	if _, err := os.Stat(realDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("password store directory does not exist: %s", realDir)
	}

	var entries []string

	var walkFunc func(currentDir string) error
	walkFunc = func(currentDir string) error {
		items, err := os.ReadDir(currentDir)
		if err != nil {
			return err
		}

		for _, item := range items {
			name := item.Name()

			// Skip hidden files or directories (.git, .gpg-id, .extensions)
			if strings.HasPrefix(name, ".") {
				continue
			}

			itemPath := filepath.Join(currentDir, name)

			// Resolve symlinks
			info, err := os.Stat(itemPath)
			if err != nil {
				continue
			}

			if info.IsDir() {
				if err := walkFunc(itemPath); err != nil {
					return err
				}
			} else if strings.HasSuffix(name, ".gpg") {
				rel, err := filepath.Rel(realDir, itemPath)
				if err != nil {
					continue
				}
				entryName := strings.TrimSuffix(rel, ".gpg")
				entries = append(entries, entryName)
			}
		}
		return nil
	}

	if err := walkFunc(realDir); err != nil {
		return nil, fmt.Errorf("failed to walk password store directory: %w", err)
	}

	sort.Strings(entries)
	return entries, nil
}

// GetEntry decrypts and parses a specific pass entry.
func (s *Store) GetEntry(name string) (*Entry, error) {
	realDir, err := filepath.EvalSymlinks(s.Dir)
	if err != nil {
		realDir = s.Dir
	}

	cleanName := strings.TrimSuffix(name, ".gpg")
	filePath := filepath.Join(realDir, cleanName+".gpg")

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
