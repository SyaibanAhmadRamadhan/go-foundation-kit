package hash_test

import (
	"testing"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/security/hash"
)

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	password := "supersecret123"

	hashPassword, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ok, err := hash.VerifyPassword(hashPassword, password)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !ok {
		t.Error("expected password to match hash, but it did not")
	}

	ok, err = hash.VerifyPassword(hashPassword, "wrongpassword")
	if err != nil {
		t.Fatalf("VerifyPassword failed (negative): %v", err)
	}
	if ok {
		t.Error("expected password not to match hash, but it did")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkpassword"
	for b.Loop() {
		_, err := hash.HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword failed: %v", err)
		}
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkpassword"
	hashPassword, err := hash.HashPassword(password)
	if err != nil {
		b.Fatalf("setup HashPassword failed: %v", err)
	}

	for b.Loop() {
		ok, err := hash.VerifyPassword(hashPassword, password)
		if err != nil {
			b.Fatalf("VerifyPassword failed: %v", err)
		}
		if !ok {
			b.Fatal("VerifyPassword returned false")
		}
	}
}
