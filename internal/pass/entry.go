package pass

import (
	"bufio"
	"net/url"
	"strings"
)

// Entry represents a decrypted pass file containing a password and optional metadata.
type Entry struct {
	Name      string            // Name/Path of the pass entry (e.g. github.com/user)
	Password  string            // Primary password
	Username  string            // Extracted username/login/email
	URL       string            // Extracted URL/URI
	OTPSecret string            // Extracted TOTP secret or otpauth:// URI
	Notes     string            // Additional unparsed notes/lines
	Custom    map[string]string // Key-value metadata
}

// ParseEntryContent parses raw decrypted pass file content into an Entry struct.
func ParseEntryContent(name string, content string) *Entry {
	entry := &Entry{
		Name:   name,
		Custom: make(map[string]string),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	var notesLines []string

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if lineNum == 1 {
			// Line 1 is the primary password by default
			entry.Password = strings.TrimSpace(line)
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for otpauth:// URI
		if strings.HasPrefix(trimmed, "otpauth://") {
			entry.OTPSecret = trimmed
			continue
		}

		// Check for Key: Value format
		if idx := strings.Index(trimmed, ":"); idx > 0 {
			key := strings.ToLower(strings.TrimSpace(trimmed[:idx]))
			val := strings.TrimSpace(trimmed[idx+1:])

			switch key {
			case "password", "pass":
				if entry.Password == "" {
					entry.Password = val
				} else {
					entry.Custom[key] = val
				}
			case "username", "user", "login", "email":
				if entry.Username == "" {
					entry.Username = val
				} else {
					entry.Custom[key] = val
				}
			case "url", "uri", "domain", "host", "website":
				if entry.URL == "" {
					entry.URL = val
				} else {
					entry.Custom[key] = val
				}
			case "otp", "totp", "secret", "2fa":
				if entry.OTPSecret == "" {
					entry.OTPSecret = val
				} else {
					entry.Custom[key] = val
				}
			default:
				entry.Custom[key] = val
			}
			continue
		}

		// Unformatted text line -> add to Notes
		notesLines = append(notesLines, line)
	}

	if len(notesLines) > 0 {
		entry.Notes = strings.Join(notesLines, "\n")
	}

	return entry
}

// ExtractOTPSecretKey returns the raw Base32 secret key from OTPSecret field or otpauth:// URI.
func (e *Entry) ExtractOTPSecretKey() string {
	if e.OTPSecret == "" {
		return ""
	}

	// If full otpauth:// URI
	if strings.HasPrefix(e.OTPSecret, "otpauth://") {
		u, err := url.Parse(e.OTPSecret)
		if err == nil {
			secret := u.Query().Get("secret")
			if secret != "" {
				return secret
			}
		}
	}

	// Raw secret key
	return strings.TrimSpace(e.OTPSecret)
}
