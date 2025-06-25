package symmetric_test

import (
	"strings"
	"testing"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/security/symmetric"
)

const ctrTestKey = "thisis32bitlongpassphraseimusing"

func TestEncryptionDecryptionCTR(t *testing.T) {
	plaintext := "ini pesan super rahasia dan sensitif"

	encrypted, err := symmetric.EncryptionStringCTR(plaintext, ctrTestKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := symmetric.DecryptionStringCTR(encrypted, ctrTestKey)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptionWithInvalidKeyLength(t *testing.T) {
	_, err := symmetric.EncryptionStringCTR("pesan", strings.Repeat("k", 33)) // >32
	if err == nil {
		t.Error("Expected error due to key length > 32, got nil")
	}
}

func TestDecryptionInvalidBase64(t *testing.T) {
	_, err := symmetric.DecryptionStringCTR("!!!not-valid-base64!!!", ctrTestKey)
	if err == nil {
		t.Error("Expected base64 decode error, got nil")
	}
}

func TestDecryptionTooShort(t *testing.T) {
	shortCipher := "dGVzdA==" // base64 dari "test", < 16 byte
	_, err := symmetric.DecryptionStringCTR(shortCipher, ctrTestKey)
	if err == nil {
		t.Error("Expected cipher too short error, got nil")
	}
}

func BenchmarkEncryptionCTR(b *testing.B) {
	data := strings.Repeat("rahasia-", 20)
	for i := 0; i < b.N; i++ {
		_, _ = symmetric.EncryptionStringCTR(data, ctrTestKey)
	}
}

func BenchmarkDecryptionCTR(b *testing.B) {
	data := strings.Repeat("rahasia-", 20)
	enc, _ := symmetric.EncryptionStringCTR(data, ctrTestKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = symmetric.DecryptionStringCTR(enc, ctrTestKey)
	}
}
