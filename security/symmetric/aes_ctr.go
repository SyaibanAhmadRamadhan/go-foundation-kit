package symmetric

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

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
