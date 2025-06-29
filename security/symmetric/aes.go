package symmetric

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"io"
)

// getIV generates a new 16-byte (AES block size) initialization vector (IV)
// using a cryptographically secure random number generator.
//
// Returns:
//   - a byte slice containing the IV
//   - an error if reading from the random source fails
func getIV() ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	return iv, nil
}

// getCipherTextAndIV separates the IV and cipher text from a single byte slice,
// assuming the IV is prepended to the cipher text.
//
// Parameters:
//   - encBytes: a byte slice containing the IV followed by the cipher text
//
// Returns:
//   - cipherText: the actual encrypted message (excluding the IV)
//   - iv: the initialization vector used during encryption (first 16 bytes)
//
// Notes:
//   - It assumes that encBytes is at least 16 bytes long (AES block size).
func getCipherTextAndIV(encBytes []byte) (cipherText []byte, iv []byte) {
	return encBytes[aes.BlockSize:], encBytes[:aes.BlockSize]
}

// paddingPKCS7 applies PKCS#7 padding to a plaintext byte slice to make its
// length a multiple of the AES block size (16 bytes).
//
// Parameters:
//   - plainText: the original unpadded byte slice
//
// Returns:
//   - a new byte slice with PKCS#7 padding appended
func paddingPKCS7(plainText []byte) []byte {
	padding := aes.BlockSize - len(plainText)%aes.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(plainText, padText...)
}

// unPaddingPKCS7 removes PKCS#7 padding from a padded plaintext byte slice.
//
// Parameters:
//   - plainText: the padded byte slice
//
// Returns:
//   - a new byte slice with padding removed
//
// Notes:
//   - Assumes the input has valid PKCS#7 padding.
//   - No padding validation is performed; use with trusted input only.
func unPaddingPKCS7(plainText []byte) []byte {
	length := len(plainText)
	unPadding := int(plainText[length-1])

	return plainText[:(length - unPadding)]
}
