package jwtx

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateHS256 creates a JWT token signed with HS256 algorithm using the provided secret and claims.
//
// Parameters:
//   - secret: the secret key used to sign the token.
//   - claims: the claims to embed in the JWT.
//
// Returns:
//   - token string if successful
//   - error if signing fails
func GenerateHS256(secret string, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// VerifyToken parses and validates a JWT token using the provided secret and expected claims.
//
// Parameters:
//   - tokenString: the JWT token string to validate.
//   - secret: the secret key used to verify the token signature.
//   - claims: a jwt.Claims instance where parsed claims will be stored.
//
// Returns:
//   - nil if the token is valid
//   - error if parsing or verification fails
func VerifyToken(tokenString string, secret string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}
