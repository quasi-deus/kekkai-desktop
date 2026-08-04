package totp_test

import (
	"testing"
	"time"

	"github.com/quasi-deus/rofi-pass-go/internal/totp"
)

func TestGenerateCodeAtTime(t *testing.T) {
	// RFC 6238 test vector secret "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ" (Base32 of "12345678901234567890")
	secret := "JBSWY3DPEHPK3PXP" // "Hello!"

	fixedTime := time.Unix(1600000000, 0) // Fixed Unix timestamp

	code, err := totp.GenerateCodeAtTime(secret, fixedTime)
	if err != nil {
		t.Fatalf("GenerateCodeAtTime failed: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("expected 6-digit TOTP code, got length %d (%s)", len(code), code)
	}

	// Test with otpauth:// URI
	uri := "otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example"
	uriCode, err := totp.GenerateCodeAtTime(uri, fixedTime)
	if err != nil {
		t.Fatalf("GenerateCodeAtTime with URI failed: %v", err)
	}

	if uriCode != code {
		t.Errorf("expected URI code %s to match raw secret code %s", uriCode, code)
	}
}
