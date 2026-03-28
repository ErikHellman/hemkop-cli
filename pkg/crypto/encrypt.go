package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/pbkdf2"
)

// Encrypt encrypts a plaintext string using the same algorithm as hemkop.se's
// frontend login form. It returns a random key and the encrypted string.
//
// The algorithm:
//  1. Generate a random 16-digit decimal key string
//  2. Generate 16 random bytes each for IV and salt
//  3. Derive an AES-128 key via PBKDF2(password=key, salt=salt, iter=1000, hash=SHA-1)
//  4. Encrypt with AES-128-CBC + PKCS7 padding
//  5. Return key and base64(hex(iv) + "::" + hex(salt) + "::" + base64(ciphertext))
func Encrypt(plaintext string) (key string, encrypted string, err error) {
	iv := make([]byte, 16)
	salt := make([]byte, 16)
	if _, err = rand.Read(iv); err != nil {
		return "", "", fmt.Errorf("generating iv: %w", err)
	}
	if _, err = rand.Read(salt); err != nil {
		return "", "", fmt.Errorf("generating salt: %w", err)
	}

	key, err = randomKey()
	if err != nil {
		return "", "", fmt.Errorf("generating key: %w", err)
	}

	encrypted, err = encryptWithParams(plaintext, key, iv, salt)
	return key, encrypted, err
}

// encryptWithParams is the deterministic core, exposed for testing.
func encryptWithParams(plaintext, key string, iv, salt []byte) (string, error) {
	derived := pbkdf2.Key([]byte(key), salt, 1000, 16, sha1.New)

	block, err := aes.NewCipher(derived)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	padded := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	inner := hex.EncodeToString(iv) + "::" + hex.EncodeToString(salt) + "::" + base64.StdEncoding.EncodeToString(ciphertext)
	return base64.StdEncoding.EncodeToString([]byte(inner)), nil
}

// randomKey generates a 16-digit decimal string, matching the JS:
// "".concat(Math.random()).substring(2,10) + "".concat(Math.random()).substring(2,10)
func randomKey() (string, error) {
	// Generate two 8-digit random numbers and concatenate
	max := new(big.Int).SetInt64(100_000_000)
	a, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	b, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d%08d", a.Int64(), b.Int64()), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := make([]byte, padding)
	for i := range pad {
		pad[i] = byte(padding)
	}
	return append(data, pad...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}

// Decrypt reverses the encryption, used for testing round-trips.
func Decrypt(encrypted, key string) (string, error) {
	outer, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("decoding outer base64: %w", err)
	}

	// Parse "hex(iv)::hex(salt)::base64(ciphertext)"
	parts := splitParts(string(outer))
	if len(parts) != 3 {
		return "", fmt.Errorf("expected 3 parts separated by ::, got %d", len(parts))
	}

	iv, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("decoding iv hex: %w", err)
	}
	salt, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decoding salt hex: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("decoding ciphertext base64: %w", err)
	}

	derived := pbkdf2.Key([]byte(key), salt, 1000, 16, sha1.New)
	block, err := aes.NewCipher(derived)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("unpadding: %w", err)
	}

	return string(unpadded), nil
}

func splitParts(s string) []string {
	var parts []string
	for {
		idx := indexOf(s, "::")
		if idx == -1 {
			parts = append(parts, s)
			break
		}
		parts = append(parts, s[:idx])
		s = s[idx+2:]
	}
	return parts
}

func indexOf(s, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
