package symmetric

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// DecryptionStringCBC decrypts a base64-encoded cipher text string that was encrypted
// using AES in CBC mode with PKCS7 padding. The IV is expected to be prefixed to the cipher text.
//
// Parameters:
//   - enc: the base64-encoded encrypted string with IV prefix
//   - key: the secret key used during encryption (must not exceed 32 characters)
//
// Returns:
//   - the decrypted plaintext string
//   - an error if decryption fails (e.g., invalid key, bad padding, malformed input)
//
// Note:
//   - The same key used for encryption must be used here.
//   - The IV is extracted from the beginning of the cipher text (first 16 bytes).
func DecryptionStringCBC(enc string, key string) (string, error) {
	if len(key) > 32 {
		return "", fmt.Errorf("%s", "key must be 32 character")
	}

	encBytes, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	if len(encBytes) < aes.BlockSize {
		return "", fmt.Errorf("cipher text too short")
	}

	cipherText, iv := getCipherTextAndIV(encBytes)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(cipherText)%aes.BlockSize != 0 {
		return "", fmt.Errorf("cipher text is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(cipherText, cipherText)

	cipherText = unPaddingPKCS7(cipherText)
	return string(cipherText), nil
}

// EncryptionStringCBC encrypts a plaintext string using AES encryption in CBC mode
// with PKCS7 padding. The result is base64-encoded and includes the IV as a prefix.
//
// Parameters:
//   - text: the plaintext string to encrypt
//   - key: the secret key (must not exceed 32 characters; padded internally)
//
// Returns:
//   - a base64-encoded encrypted string that includes the IV prefix
//   - an error if encryption fails or the key is too long
//
// Note:
//   - The key should be 16, 24, or 32 bytes long for AES-128, AES-192, or AES-256 respectively.
//   - If the key is shorter than 32 bytes, AES will pad/truncate internally.
//   - The IV is randomly generated per encryption and prepended to the cipher text.
func EncryptionStringCBC(text string, key string) (string, error) {
	if len(key) > 32 {
		return "", fmt.Errorf("%s", "key must be 32 character")
	}

	iv, err := getIV()
	if err != nil {
		return "", err
	}

	plainText := []byte(text)
	plainText = paddingPKCS7(plainText)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, len(plainText))

	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(cipherText, plainText)

	final := append(iv, cipherText...)
	return base64.StdEncoding.EncodeToString(final), nil
}
