package domain

import (
	"backend/internal/config"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// issueAccessToken mints a short-lived JWT for a user.
func issueAccessToken(userId, role string) (string, error) {
	expiry := time.Duration(config.Int("authentication.accessTokenExpiryMinutes")) * time.Minute
	claims := AccessTokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.String("authentication.jwtSecret")))
}

// ValidateAccessToken verifies the JWT's signature and expiry.
func ValidateAccessToken(rawToken string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(rawToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(config.String("authentication.jwtSecret")), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// generateRefreshToken returns a random token and its SHA-256 hash for storage.
func generateRefreshToken() (rawToken, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}

	rawToken = base64.RawURLEncoding.EncodeToString(buf)
	hash = hashRefreshToken(rawToken)
	return rawToken, hash, nil
}

func hashRefreshToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
