// Package repository defines the contract connectors use to read secrets,
// independent of which backend (pass today; Bitwarden, KeePass, etc. later)
// provides them.
package repository

// Repository lists and reads entries from a secret-store backend.
type Repository interface {
	ListEntries() ([]string, error)
	GetPassword(name string) (string, error)
}
