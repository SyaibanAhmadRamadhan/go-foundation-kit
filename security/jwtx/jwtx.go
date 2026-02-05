package jwtx

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Error variables untuk JWT operations
var (
	ErrEmptyToken          = errors.New("empty token")   // error ketika token string kosong
	ErrExpiredToken        = errors.New("expired token") // error ketika token sudah kadaluarsa
	ErrInvalidToken        = errors.New("invalid token") // error ketika token tidak valid atau malformed
	ErrServiceUnauthorized = errors.New("Service Unauthorized")
)

// EzClaims adalah struktur custom claims untuk JWT token yang menggabungkan
// data kustom dengan registered claims standar JWT.
type EzClaims struct {
	Data                 map[string]any `json:"data,omitempty"` // data kustom yang ingin disimpan dalam token
	jwt.RegisteredClaims                // claims standar JWT (iss, sub, exp, iat, dll)
}

// RespToken adalah struktur response yang dikembalikan setelah token dibuat.
// Berisi access token yang sudah dienkripsi beserta informasi expiry.
type RespToken struct {
	Token     string
	ExpiredAt time.Time // waktu kadaluarsa token dalam format datetime
}

// Signer adalah interface yang mendefinisikan kontrak untuk signing dan verifikasi JWT.
// Interface ini mem-enforce algoritma signing dan menyediakan key untuk verifikasi/signing.
type Signer interface {
	Method() jwt.SigningMethod               // mengembalikan metode signing yang digunakan
	Sign(token *jwt.Token) (string, error)   // melakukan signing terhadap token
	VerifyKey(token *jwt.Token) (any, error) // mengembalikan key untuk verifikasi token
}

// JWT adalah struktur utama yang menangani enkripsi, signing, dan parsing JWT token.
// JWT token akan di-sign terlebih dahulu, kemudian dienkripsi untuk keamanan tambahan.
type JWT struct {
	issuer string           // issuer token (siapa yang mengeluarkan token)
	signer Signer           // signer untuk signing dan verifikasi JWT
	now    func() time.Time // fungsi untuk mendapatkan waktu sekarang (injectable untuk testing)
}

// Option adalah function type untuk konfigurasi opsional JWT
type Option func(*JWT)

// WithIssuer adalah option function untuk mengatur issuer JWT token.
// Issuer menandakan siapa yang mengeluarkan token tersebut.
//
// Parameters:
//   - issuer: nama issuer (contoh: "MY-SERVICE-API")
//
// Returns:
//   - Option: function yang akan mengatur issuer pada JWT instance
func WithIssuer(issuer string) Option {
	return func(j *JWT) { j.issuer = issuer }
}

// WithNow adalah option function untuk mengatur fungsi penyedia waktu.
// Berguna untuk testing dengan waktu yang bisa dikontrol.
//
// Parameters:
//   - fn: fungsi yang mengembalikan waktu sekarang
//
// Returns:
//   - Option: function yang akan mengatur time provider pada JWT instance
func WithNow(fn func() time.Time) Option {
	return func(j *JWT) { j.now = fn }
}

// New membuat instance baru JWT dengan konfigurasi yang diberikan.
// JWT akan menggunakan enkripsi AES untuk melindungi token string dan
// signer untuk signing/verifikasi token.
//
// Parameters:
//   - keyEncrypt: key untuk enkripsi/dekripsi token string menggunakan AES-CFB (required jika encrypted=true)
//   - signer: implementasi Signer untuk signing dan verifikasi token
//   - opts: option functions untuk konfigurasi tambahan (issuer, now function, encrypted flag, dll)
//
// Returns:
//   - *JWT: instance JWT yang siap digunakan
//
// Panics:
//   - jika encrypted=true (default) dan keyEncrypt kosong
func New(keyEncrypt string, signer Signer, opts ...Option) *JWT {
	j := &JWT{
		issuer: "PLATFORM-AUTH-API",
		signer: signer,
		now:    time.Now().UTC,
	}
	for _, o := range opts {
		o(j)
	}

	return j
}

// CreateToken membuat JWT token baru dengan data dan konfigurasi yang diberikan.
// Token akan berisi data kustom, subject, issuer, issued at, dan expiry time.
//
// Parameters:
//   - data: map berisi data kustom yang ingin disimpan dalam token
//   - subject: subject token (biasanya user ID atau identifier)
//   - ttlSeconds: time-to-live token dalam detik (berapa lama token valid)
//
// Returns:
//   - RespToken: response berisi access token yang sudah dienkripsi dan waktu expiry
//   - error: error jika terjadi kesalahan saat signing atau enkripsi
func (j *JWT) CreateToken(data map[string]any, subject string, ttlSeconds int64) (RespToken, error) {
	now := j.now()
	exp := now.Add(time.Duration(ttlSeconds) * time.Second)

	claims := EzClaims{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	tok := jwt.NewWithClaims(j.signer.Method(), claims)

	signed, err := j.signer.Sign(tok)
	if err != nil {
		return RespToken{}, err
	}

	return RespToken{
		Token:     signed,
		ExpiredAt: exp,
	}, nil
}

// GenerateToken membuat JWT token dari claims yang sudah diberikan.
// Method ini lebih fleksibel karena menerima EzClaims yang sudah dikonfigurasi,
// memungkinkan kontrol penuh atas semua fields dalam token.
//
// Parameters:
//   - data: EzClaims berisi semua claims yang ingin disimpan dalam token
//
// Returns:
//   - string: JWT token yang sudah di-sign dan dienkripsi
//   - error: error jika terjadi kesalahan saat signing atau enkripsi
func (j *JWT) GenerateToken(data EzClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	signed, err := j.signer.Sign(token)
	if err != nil {
		return "", err
	}

	// Skip enkripsi jika encrypted=false
	return signed, nil
}

// ValidateToken memvalidasi JWT token yang diberikan.
// Method ini akan:
// 1. Mendekripsi token string menggunakan AES
// 2. Parse JWT dan validasi signature menggunakan signer
// 3. Validasi expiry time dan format token
// 4. Mengembalikan parsed token jika valid
//
// Parameters:
//   - tokenString: JWT token yang sudah dienkripsi (hasil dari CreateToken/GenerateToken)
//
// Returns:
//   - *jwt.Token: parsed JWT token jika valid
//   - error: error spesifik (ErrEmptyToken, ErrExpiredToken, atau ErrInvalidToken)
func (j *JWT) ValidateToken(tokenString string) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, ErrEmptyToken
	}

	// Parse + enforce expected signing method
	tok, err := jwt.ParseWithClaims(tokenString, &EzClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != j.signer.Method().Alg() {
			return nil, fmt.Errorf("%w: unexpected alg %s", ErrInvalidToken, t.Method.Alg())
		}
		return j.signer.VerifyKey(t)
	})

	if err == nil {
		return tok, nil
	}

	switch {
	case errors.Is(err, jwt.ErrTokenMalformed):
		return nil, ErrInvalidToken
	case errors.Is(err, jwt.ErrTokenExpired), errors.Is(err, jwt.ErrTokenNotValidYet):
		return nil, ErrExpiredToken
	default:
		return nil, ErrInvalidToken
	}
}

// GetTokenData mengekstrak claims dari JWT token yang diberikan.
// Method ini adalah shortcut untuk ValidateToken + ekstraksi claims,
// mengembalikan EzClaims jika token valid.
//
// Parameters:
//   - token: JWT token yang sudah dienkripsi
//
// Returns:
//   - *EzClaims: claims yang berisi data dan registered claims dari token
//   - error: error jika token tidak valid atau gagal divalidasi
func (j *JWT) GetTokenData(token string) (*EzClaims, error) {
	tok, err := j.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*EzClaims)
	if !ok || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
