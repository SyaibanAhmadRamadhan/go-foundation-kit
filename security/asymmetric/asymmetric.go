package asymmetric

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// PrivateKeyToBytes encodes a private key into PEM format bytes using PKCS#8.
//
// Parameters:
//   - priv: a private key (e.g., *rsa.PrivateKey, *ecdsa.PrivateKey).
//
// Returns:
//   - PEM-encoded private key as []byte
//   - error if marshaling fails
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

// PublicKeyToBytes encodes a public key into PEM format bytes.
//
// Parameters:
//   - pub: a public key (e.g., *rsa.PublicKey, *ecdsa.PublicKey).
//
// Returns:
//   - PEM-encoded public key as []byte
//   - error if marshaling fails
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

// BytesToPrivateKeyGeneric decodes PEM-formatted private key bytes into a concrete private key type.
//
// Type Parameters:
//   - T: the expected private key type (e.g., *rsa.PrivateKey, *ecdsa.PrivateKey).
//
// Parameters:
//   - pemBytes: the PEM-encoded private key bytes.
//
// Returns:
//   - the decoded private key of type T
//   - error if decoding or type assertion fails
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

// BytesToPublicKeyGeneric decodes PEM-formatted public key bytes into a concrete public key type.
//
// Type Parameters:
//   - T: the expected public key type (e.g., *rsa.PublicKey, *ecdsa.PublicKey).
//
// Parameters:
//   - pemBytes: the PEM-encoded public key bytes.
//
// Returns:
//   - the decoded public key of type T
//   - error if decoding or type assertion fails
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

// Base64ToPrivateKeyGeneric decodes a base64-encoded private key string into a concrete private key type.
//
// Type Parameters:
//   - T: the expected private key type (e.g., *rsa.PrivateKey, *ecdsa.PrivateKey).
//
// Parameters:
//   - key: the base64-encoded private key string.
//
// Returns:
//   - the decoded private key of type T
//   - error if decoding or type assertion fails
func Base64ToPrivateKeyGeneric[T any](key string) (T, error) {
	var zero T

	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return zero, err
	}

	return BytesToPrivateKeyGeneric[T](decoded)
}

// Base64ToPublicKeyGeneric decodes a base64-encoded public key string into a concrete public key type.
//
// Type Parameters:
//   - T: the expected public key type (e.g., *rsa.PublicKey, *ecdsa.PublicKey).
//
// Parameters:
//   - key: the base64-encoded public key string.
//
// Returns:
//   - the decoded public key of type T
//   - error if decoding or type assertion fails
func Base64ToPublicKeyGeneric[T any](key string) (T, error) {
	var zero T

	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return zero, err
	}

	return BytesToPublicKeyGeneric[T](decoded)
}
