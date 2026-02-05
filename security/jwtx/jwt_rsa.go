package jwtx

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// RSASigner adalah implementasi Signer yang menggunakan algoritma RSA
// untuk signing dan verifikasi JWT token. RSA menggunakan asymmetric cryptography
// di mana private key digunakan untuk signing dan public key untuk verifikasi.
type RSASigner struct {
	privateKey *rsa.PrivateKey // private key RSA untuk signing token
	publicKey  *rsa.PublicKey  // public key RSA untuk verifikasi token
}

// NewRS256Signer membuat instance baru RSASigner dengan algoritma RS256.
// RS256 menggunakan RSA signature dengan SHA-256 hash.
//
// Parameters:
//   - priv: RSA private key untuk signing token
//   - pub: RSA public key untuk verifikasi token
//
// Returns:
//   - *RSASigner: instance RSASigner yang siap digunakan
func NewRS256Signer(priv *rsa.PrivateKey, pub *rsa.PublicKey) *RSASigner {
	return &RSASigner{
		privateKey: priv,
		publicKey:  pub,
	}
}

// Method mengembalikan metode signing JWT yang digunakan oleh RSASigner ini.
// Method ini adalah bagian dari interface Signer dan selalu mengembalikan RS256.
//
// Returns:
//   - jwt.SigningMethod: jwt.SigningMethodRS256
func (s *RSASigner) Method() jwt.SigningMethod {
	return jwt.SigningMethodRS256
}

// Sign melakukan signing terhadap JWT token menggunakan RSA private key.
// Method ini akan mem-force token untuk menggunakan RS256 algorithm,
// kemudian membuat signed string menggunakan RSA private key.
//
// Parameters:
//   - token: JWT token yang akan di-sign
//
// Returns:
//   - string: JWT token yang sudah di-sign dalam format string
//   - error: error jika terjadi kesalahan saat signing
func (s *RSASigner) Sign(token *jwt.Token) (string, error) {
	token.Method = jwt.SigningMethodRS256
	return token.SignedString(s.privateKey)
}

// VerifyKey mengembalikan public key yang digunakan untuk verifikasi JWT token.
// Method ini memvalidasi bahwa algorithm token adalah RS256,
// kemudian mengembalikan RSA public key untuk proses verifikasi.
//
// Parameters:
//   - token: JWT token yang akan diverifikasi
//
// Returns:
//   - any: RSA public key untuk verifikasi
//   - error: error jika algorithm token bukan RS256
func (s *RSASigner) VerifyKey(token *jwt.Token) (any, error) {
	if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
		return nil, fmt.Errorf("unexpected alg: %s", token.Method.Alg())
	}
	return s.publicKey, nil
}
