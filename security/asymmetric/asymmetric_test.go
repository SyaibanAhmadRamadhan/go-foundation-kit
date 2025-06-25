package asymmetric_test

import (
	"bytes"
	"crypto/rsa"
	"testing"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/security/asymmetric"
)

func TestRSAEncryptDecrypt(t *testing.T) {
	// Generate keypair
	privKey, pubKey, err := asymmetric.RSAGenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	// Serialize private key
	privPem, err := asymmetric.PrivateKeyToBytes(privKey)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}

	// Serialize public key
	pubPem, err := asymmetric.PublicKeyToBytes(pubKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}

	// Deserialize private key
	parsedPriv, err := asymmetric.BytesToPrivateKeyGeneric[*rsa.PrivateKey](privPem)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	// Deserialize public key
	parsedPub, err := asymmetric.BytesToPublicKeyGeneric[*rsa.PublicKey](pubPem)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	// Encrypt with public key
	msg := []byte("hello world")
	enc, err := asymmetric.RSAEncryptWithPublicKey(msg, parsedPub)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decrypt with private key
	dec, err := asymmetric.RSADecryptWithPrivateKey(enc, parsedPriv)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Validate decrypted message
	if !bytes.Equal(msg, dec) {
		t.Errorf("decrypted message mismatch: got %s, want %s", dec, msg)
	}
}

func BenchmarkRSAEncryptDecrypt(b *testing.B) {
	privKey, pubKey, _ := asymmetric.RSAGenerateKeyPair(2048)
	msg := []byte("benchmark testing")

	for i := 0; i < b.N; i++ {
		enc, err := asymmetric.RSAEncryptWithPublicKey(msg, pubKey)
		if err != nil {
			b.Fatal(err)
		}
		_, err = asymmetric.RSADecryptWithPrivateKey(enc, privKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}
