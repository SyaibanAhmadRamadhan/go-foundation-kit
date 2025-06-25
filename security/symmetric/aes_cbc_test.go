package symmetric_test

import (
	"strings"
	"testing"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/security/symmetric"
)

const testKey = "thisis32bitlongpassphraseimusing"

func TestEncryptionDecryptionCBC(t *testing.T) {
	plaintext := "Ini adalah pesan rahasia super penting."

	// Encrypt
	cipherText, err := symmetric.EncryptionStringCBC(plaintext, testKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Decrypt
	result, err := symmetric.DecryptionStringCBC(cipherText, testKey)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if result != plaintext {
		t.Errorf("Expected %q but got %q", plaintext, result)
	}
}

func TestEncryptionWithInvalidKey(t *testing.T) {
	_, err := symmetric.EncryptionStringCBC("pesan", strings.Repeat("x", 33)) // >32 char
	if err == nil {
		t.Error("Expected error due to long key, got nil")
	}
}

func TestDecryptionWithInvalidBase64(t *testing.T) {
	_, err := symmetric.DecryptionStringCBC("!!not-base64!!", testKey)
	if err == nil {
		t.Error("Expected base64 decode error, got nil")
	}
}

func TestDecryptionWithTooShortCipher(t *testing.T) {
	// base64 of less than 16 bytes
	shortBase64 := "dGVzdA==" // "test"
	_, err := symmetric.DecryptionStringCBC(shortBase64, testKey)
	if err == nil {
		t.Error("Expected cipher too short error, got nil")
	}
}

func BenchmarkEncryptionCBC(b *testing.B) {
	data := strings.Repeat("data-rahasia-", 10) // ~140 bytes
	for i := 0; i < b.N; i++ {
		_, _ = symmetric.EncryptionStringCBC(data, testKey)
	}
}

func BenchmarkDecryptionCBC(b *testing.B) {
	data := strings.Repeat("data-rahasia-", 10)
	cipherText, _ := symmetric.EncryptionStringCBC(data, testKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = symmetric.DecryptionStringCBC(cipherText, testKey)
	}
}
