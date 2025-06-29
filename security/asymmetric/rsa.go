package asymmetric

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
)

// RSAGenerateKeyPair generates a new RSA key pair.
//
// Parameters:
//   - bits: key size in bits (e.g., 2048, 3072, 4096)
//
// Returns:
//   - *rsa.PrivateKey: the generated private key
//   - *rsa.PublicKey: the corresponding public key
//   - error: any error encountered during generation
func RSAGenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privkey, &privkey.PublicKey, nil
}

// RSAEncryptWithPublicKey encrypts a message using the recipient's RSA public key and OAEP padding with SHA-512.
//
// Parameters:
//   - msg: the plaintext message as []byte
//   - pub: the recipient's RSA public key
//
// Returns:
//   - ciphertext: the encrypted message as []byte
//   - error: any error encountered during encryption
func RSAEncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha512.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// RSADecryptWithPrivateKey decrypts an RSA-encrypted message using the private key and OAEP padding with SHA-512.
//
// Parameters:
//   - ciphertext: the encrypted message as []byte
//   - priv: the RSA private key
//
// Returns:
//   - plaintext: the decrypted message as []byte
//   - error: any error encountered during decryption
func RSADecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
