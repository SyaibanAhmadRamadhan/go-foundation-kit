package symmetric

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"io"
)

func getIV() ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	return iv, nil
}

func getCipherTextAndIV(encBytes []byte) (cipherText []byte, iv []byte) {
	return encBytes[aes.BlockSize:], encBytes[:aes.BlockSize]
}

func paddingPKCS7(plainText []byte) []byte {
	padding := aes.BlockSize - len(plainText)%aes.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(plainText, padText...)
}

func unPaddingPKCS7(plainText []byte) []byte {
	length := len(plainText)
	unPadding := int(plainText[length-1])

	return plainText[:(length - unPadding)]
}
