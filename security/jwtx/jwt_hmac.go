package jwtx

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// HMACSigner adalah implementasi Signer yang menggunakan algoritma HMAC
// untuk signing dan verifikasi JWT token. HMAC (Hash-based Message Authentication Code)
// menggunakan secret key yang sama untuk signing dan verifikasi.
type HMACSigner struct {
	method jwt.SigningMethod // metode signing JWT yang digunakan (contoh: HS256, HS384, HS512)
	secret []byte            // secret key dalam bentuk byte array untuk signing dan verifikasi
}

// NewHMACSigner membuat instance baru HMACSigner dengan konfigurasi yang diberikan.
//
// Parameters:
//   - method: metode signing JWT (contoh: jwt.SigningMethodHS256)
//   - secret: secret key dalam format string (bisa terenkripsi atau plain text)
//
// Returns:
//   - *HMACSigner: instance HMACSigner yang siap digunakan
func NewHMACSigner(method jwt.SigningMethod, secret string) *HMACSigner {
	return &HMACSigner{
		method: method,
		secret: []byte(secret),
	}
}

// Method mengembalikan metode signing JWT yang digunakan oleh HMACSigner ini.
// Method ini adalah bagian dari interface Signer.
//
// Returns:
//   - jwt.SigningMethod: metode signing yang dikonfigurasi (contoh: HS256)
func (s *HMACSigner) Method() jwt.SigningMethod { return s.method }

// Sign melakukan signing terhadap JWT token menggunakan HMAC algorithm.
// Method ini akan mem-force token untuk menggunakan signing method yang telah
// dikonfigurasi, kemudian membuat signed string menggunakan secret key.
//
// Parameters:
//   - token: JWT token yang akan di-sign
//
// Returns:
//   - string: JWT token yang sudah di-sign dalam format string
//   - error: error jika terjadi kesalahan saat signing
func (s *HMACSigner) Sign(token *jwt.Token) (string, error) {
	// enforce method
	token.Method = s.method
	return token.SignedString(s.secret)
}

// VerifyKey mengembalikan key yang digunakan untuk verifikasi JWT token.
// Method ini memvalidasi bahwa algorithm token sesuai dengan yang dikonfigurasi,
// kemudian mengembalikan secret key untuk proses verifikasi.
//
// Parameters:
//   - token: JWT token yang akan diverifikasi
//
// Returns:
//   - any: secret key untuk verifikasi (dalam bentuk []byte)
//   - error: error jika algorithm token tidak sesuai
func (s *HMACSigner) VerifyKey(token *jwt.Token) (any, error) {
	if token.Method.Alg() != s.method.Alg() {
		return nil, fmt.Errorf("unexpected alg: %s", token.Method.Alg())
	}
	return s.secret, nil
}
