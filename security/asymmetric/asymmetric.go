package asymmetric

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func PrivateKeyToBytes(priv any) ([]byte, error) {
	bytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}

	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: bytes,
		},
	)

	return privBytes, nil
}

func PublicKeyToBytes(pub any) ([]byte, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes, nil
}

func BytesToPrivateKeyGeneric[T any](pemBytes []byte) (T, error) {
	var zero T

	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "PRIVATE KEY" {
		return zero, fmt.Errorf("invalid PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return zero, err
	}

	casted, ok := key.(T)
	if !ok {
		return zero, fmt.Errorf("not type %T", zero)
	}

	return casted, nil
}

func BytesToPublicKeyGeneric[T any](pemBytes []byte) (T, error) {
	var zero T

	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return zero, fmt.Errorf("invalid PEM block or type")
	}

	keyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return zero, err
	}

	key, ok := keyInterface.(T)
	if !ok {
		return zero, fmt.Errorf("not type %T", zero)
	}

	return key, nil
}
