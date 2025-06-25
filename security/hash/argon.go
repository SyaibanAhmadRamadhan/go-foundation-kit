package hash

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory     = 64 * 1024 // 64 MB
	time       = 3         // Iterasi
	threads    = 4         // Paralelisme
	keyLen     = 32        // length hasil hash
	saltLength = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, uint8(threads), keyLen)

	encoded := fmt.Sprintf("%d$%d$%d$%s$%s", time, memory, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

func VerifyPassword(encodedHash, password string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return false, errors.New("invalid hash format")
	}

	var (
		time    uint32
		memory  uint32
		threads uint8
	)

	fmt.Sscanf(parts[0], "%d", &time)
	fmt.Sscanf(parts[1], "%d", &memory)
	fmt.Sscanf(parts[2], "%d", &threads)

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash)))
	return string(hash) == string(expectedHash), nil
}
