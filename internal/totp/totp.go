package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// GenerateCode computes a TOTP code from a secret or otpauth:// URI at the current time.
func GenerateCode(secretOrURI string) (string, error) {
	return GenerateCodeAtTime(secretOrURI, time.Now())
}

// GenerateCodeAtTime computes a TOTP code from a secret or otpauth:// URI for a specific time.
func GenerateCodeAtTime(secretOrURI string, t time.Time) (string, error) {
	secret := secretOrURI
	digits := 6
	period := int64(30)

	// Check if secret is an otpauth:// URI
	if strings.HasPrefix(secretOrURI, "otpauth://") {
		u, err := url.Parse(secretOrURI)
		if err == nil {
			q := u.Query()
			if s := q.Get("secret"); s != "" {
				secret = s
			}
			if d := q.Get("digits"); d != "" {
				if val, err := strconv.Atoi(d); err == nil && (val == 6 || val == 8) {
					digits = val
				}
			}
			if p := q.Get("period"); p != "" {
				if val, err := strconv.ParseInt(p, 10, 64); err == nil && val > 0 {
					period = val
				}
			}
		}
	}

	// Clean Base32 secret (uppercase, remove spaces and hyphens)
	cleanSecret := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(secret, " ", ""), "-", ""))
	
	// Add Base32 padding if missing
	if mod := len(cleanSecret) % 8; mod != 0 {
		cleanSecret += strings.Repeat("=", 8-mod)
	}

	key, err := base32.StdEncoding.DecodeString(cleanSecret)
	if err != nil {
		return "", fmt.Errorf("invalid Base32 TOTP secret: %w", err)
	}

	// RFC 6238 counter calculation
	counter := uint64(t.Unix() / period)

	// Pack counter as 8-byte big-endian integer
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	// HMAC-SHA1 calculation
	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	truncatedHash := binary.BigEndian.Uint32(hash[offset : offset+4])
	truncatedHash &= 0x7fffffff

	otp := truncatedHash % uint32(math.Pow10(digits))

	// Format with leading zeros
	return fmt.Sprintf("%0*d", digits, otp), nil
}
