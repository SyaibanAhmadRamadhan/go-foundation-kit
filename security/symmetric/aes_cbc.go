package symmetric

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

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
