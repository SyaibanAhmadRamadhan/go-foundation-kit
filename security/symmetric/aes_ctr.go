package symmetric

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// DecryptionStringCTR decrypts a base64-encoded string that was encrypted using AES in CTR mode.
// The IV (nonce) is extracted from the first 16 bytes of the decoded input.
//
// Parameters:
//   - text: the base64-encoded string containing IV + cipher text
//   - key: the secret key used during encryption (must be 32 characters or fewer)
//
// Returns:
//   - the decrypted plaintext string
//   - an error if decryption fails (e.g., malformed input, invalid key, or decryption error)
//
// Notes:
//   - The same key and IV used during encryption must be used for correct decryption.
//   - CTR mode is symmetric and stateless (no padding involved).
func DecryptionStringCTR(text string, key string) (string, error) {
	if len(key) > 32 {
		return "", fmt.Errorf("%s", "key must be 32 character")
	}

	encBytes, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	if len(encBytes) < aes.BlockSize {
		return "", fmt.Errorf("%s", "cipher text is too short")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	cipherText, iv := getCipherTextAndIV(encBytes)

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

// EncryptionStringCTR encrypts a plaintext string using AES in CTR (Counter) mode.
// The IV (nonce) is randomly generated and prepended to the encrypted output.
//
// Parameters:
//   - text: the plaintext string to encrypt
//   - key: the secret key (must be 32 characters or fewer)
//
// Returns:
//   - a base64-encoded encrypted string (IV + cipher text)
//   - an error if encryption fails (e.g., key too long, IV generation error)
//
// Notes:
//   - The key length should be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256 respectively.
//   - The generated IV is 16 bytes (AES block size) and securely random.
//   - CTR mode does not require padding, making it suitable for streaming use cases.
func EncryptionStringCTR(text string, key string) (string, error) {
	if len(key) > 32 {
		return "", fmt.Errorf("%s", "key must be 32 character")
	}

	iv, err := getIV()
	if err != nil {
		return "", err
	}

	plainText := []byte(text)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, len(plainText))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText, plainText)

	final := append(iv, cipherText...)
	return base64.StdEncoding.EncodeToString(final), nil
}
