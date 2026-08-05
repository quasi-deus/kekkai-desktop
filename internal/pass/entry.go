package pass

import (
	"bufio"
	"strings"
)

// ParsePassword extracts the primary password (first line) from decrypted GPG content.
func ParsePassword(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}
