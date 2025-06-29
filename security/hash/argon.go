package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ConfigArgon2ID holds the Argon2 hashing parameters.
type ConfigArgon2ID struct {
	Memory     uint32
	Time       uint32
	Threads    uint8
	KeyLen     uint32
	SaltLength int
}

// DefaultConfigArgon2ID provides recommended secure defaults.
var DefaultConfigArgon2ID = ConfigArgon2ID{
	Memory:     64 * 1024, // 64 MB
	Time:       3,
	Threads:    4,
	KeyLen:     32,
	SaltLength: 16,
}

// Hasher provides methods to hash and verify passwords using Argon2id.
type HasherArgon2ID struct {
	Config ConfigArgon2ID
}

// NewHasher returns a new Hasher with the given config.
func NewHasherArgon2ID(cfg ConfigArgon2ID) *HasherArgon2ID {
	return &HasherArgon2ID{Config: cfg}
}

// Hash hashes the password using Argon2id and encodes it into a string.
func (h *HasherArgon2ID) Hash(password string) (string, error) {
	salt := make([]byte, h.Config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, h.Config.Time, h.Config.Memory, h.Config.Threads, h.Config.KeyLen)

	encoded := fmt.Sprintf("%d$%d$%d$%s$%s",
		h.Config.Time,
		h.Config.Memory,
		h.Config.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// Verify compares a plain password with an encoded hash.
// Returns true if it matches, false otherwise, and error if the format is invalid.
func (h *HasherArgon2ID) Verify(encodedHash string, password string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return false, errors.New("invalid hash format")
	}

	var (
		time    uint32
		memory  uint32
		threads uint8
	)

	if _, err := fmt.Sscanf(parts[0], "%d", &time); err != nil {
		return false, err
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &memory); err != nil {
		return false, err
	}
	if _, err := fmt.Sscanf(parts[2], "%d", &threads); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash)))

	if subtle.ConstantTimeCompare(actualHash, expectedHash) == 1 {
		return true, nil
	}
	return false, nil
}
