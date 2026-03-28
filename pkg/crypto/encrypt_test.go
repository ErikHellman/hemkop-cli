package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"short", "hello"},
		{"swedish_ssn", "197702065538"},
		{"password", "kmc@VvMT@y6XyfF%DbcP"},
		{"exact_block_size", "1234567890123456"}, // exactly 16 bytes
		{"empty", ""},
		{"unicode", "åäö"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, encrypted, err := Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if len(key) != 16 {
				t.Errorf("key length = %d, want 16", len(key))
			}

			// Verify key is all digits
			for _, c := range key {
				if c < '0' || c > '9' {
					t.Errorf("key contains non-digit: %c", c)
				}
			}

			// Verify output format: base64 -> "hex::hex::base64"
			decoded, err := base64.StdEncoding.DecodeString(encrypted)
			if err != nil {
				t.Fatalf("outer base64 decode failed: %v", err)
			}
			parts := strings.SplitN(string(decoded), "::", 3)
			if len(parts) != 3 {
				t.Fatalf("expected 3 parts, got %d: %s", len(parts), string(decoded))
			}
			if len(parts[0]) != 32 { // 16 bytes = 32 hex chars
				t.Errorf("IV hex length = %d, want 32", len(parts[0]))
			}
			if len(parts[1]) != 32 {
				t.Errorf("salt hex length = %d, want 32", len(parts[1]))
			}

			// Decrypt and verify round-trip
			decrypted, err := Decrypt(encrypted, key)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}
			if decrypted != tt.plaintext {
				t.Errorf("round-trip failed: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptDeterministic(t *testing.T) {
	// With fixed parameters, encryption should produce a known output
	iv := make([]byte, 16)
	salt := make([]byte, 16)
	for i := range iv {
		iv[i] = byte(i)
	}
	for i := range salt {
		salt[i] = byte(i + 16)
	}
	key := "1234567890123456"

	encrypted, err := encryptWithParams("hello", key, iv, salt)
	if err != nil {
		t.Fatalf("encryptWithParams failed: %v", err)
	}

	// Verify we can decrypt it
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != "hello" {
		t.Errorf("got %q, want %q", decrypted, "hello")
	}

	// Run again with same params - should produce identical output
	encrypted2, err := encryptWithParams("hello", key, iv, salt)
	if err != nil {
		t.Fatalf("encryptWithParams failed: %v", err)
	}
	if encrypted != encrypted2 {
		t.Errorf("deterministic encryption produced different results")
	}
}

func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		inputLen    int
		expectedLen int
	}{
		{0, 16},
		{1, 16},
		{15, 16},
		{16, 32}, // full block -> adds another full block of padding
		{17, 32},
		{31, 32},
		{32, 48},
	}

	for _, tt := range tests {
		data := make([]byte, tt.inputLen)
		padded := pkcs7Pad(data, 16)
		if len(padded) != tt.expectedLen {
			t.Errorf("pkcs7Pad(len=%d): got len %d, want %d", tt.inputLen, len(padded), tt.expectedLen)
		}

		unpadded, err := pkcs7Unpad(padded)
		if err != nil {
			t.Errorf("pkcs7Unpad failed for inputLen=%d: %v", tt.inputLen, err)
		}
		if len(unpadded) != tt.inputLen {
			t.Errorf("pkcs7Unpad(len=%d): got len %d, want %d", tt.inputLen, len(unpadded), tt.expectedLen)
		}
	}
}
