package hashx_test

import (
	"testing"

	hashx "github.com/SyaibanAhmadRamadhan/go-foundation-kit/security/hash"
)

func TestHasher_HashAndVerify(t *testing.T) {
	hasher := hashx.NewHasherArgon2ID(hashx.DefaultConfigArgon2ID)
	password := "supersecret123"

	t.Run("success verify", func(t *testing.T) {
		hashPassword, err := hasher.Hash(password)
		if err != nil {
			t.Fatalf("Hash() failed: %v", err)
		}

		ok, err := hasher.Verify(hashPassword, password)
		if err != nil {
			t.Fatalf("Verify() failed: %v", err)
		}
		if !ok {
			t.Error("expected password to match hash, but it did not")
		}
	})

	t.Run("failed verify", func(t *testing.T) {
		hashPassword, err := hasher.Hash(password)
		if err != nil {
			t.Fatalf("Hash() failed: %v", err)
		}

		ok, err := hasher.Verify(hashPassword, "wrongpassword")
		if err != nil {
			t.Fatalf("Verify() failed (negative): %v", err)
		}
		if ok {
			t.Error("expected password not to match hash, but it did")
		}
	})

	t.Run("invalid hash format", func(t *testing.T) {
		_, err := hasher.Verify("invalid$format$only$3$parts", password)
		if err == nil {
			t.Error("expected error on invalid hash format, but got nil")
		}
	})
}

func BenchmarkHasher_Hash(b *testing.B) {
	hasher := hashx.NewHasherArgon2ID(hashx.DefaultConfigArgon2ID)
	password := "benchmarkpassword"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(password)
		if err != nil {
			b.Fatalf("Hash failed: %v", err)
		}
	}
}

func BenchmarkHasher_Verify(b *testing.B) {
	hasher := hashx.NewHasherArgon2ID(hashx.DefaultConfigArgon2ID)
	password := "benchmarkpassword"

	hashPassword, err := hasher.Hash(password)
	if err != nil {
		b.Fatalf("setup Hash failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ok, err := hasher.Verify(hashPassword, password)
		if err != nil {
			b.Fatalf("Verify failed: %v", err)
		}
		if !ok {
			b.Fatal("Verify returned false")
		}
	}
}
