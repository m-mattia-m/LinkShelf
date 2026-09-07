package domain

import (
	"backend/internal/config"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func setupJwtTestConfig(t *testing.T) {
	t.Helper()
	config.Reset()
	config.Set("authentication.jwtSecret", "test-secret")
	config.Set("authentication.accessTokenExpiryMinutes", 5)
}

func Test_Unit_IssueAndValidateAccessToken_RoundTrip(t *testing.T) {
	setupJwtTestConfig(t)

	token, err := issueAccessToken("user-uuid-test", "admin")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := ValidateAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, "user-uuid-test", claims.Subject)
	require.Equal(t, "admin", claims.Role)
}

func Test_Unit_ValidateAccessToken_Expired(t *testing.T) {
	setupJwtTestConfig(t)
	config.Set("authentication.accessTokenExpiryMinutes", 0)

	claims := AccessTokenClaims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-uuid-test",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.String("authentication.jwtSecret")))
	require.NoError(t, err)

	_, err = ValidateAccessToken(signed)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func Test_Unit_ValidateAccessToken_WrongSecret(t *testing.T) {
	setupJwtTestConfig(t)

	token, err := issueAccessToken("user-uuid-test", "user")
	require.NoError(t, err)

	config.Set("authentication.jwtSecret", "a-different-secret")

	_, err = ValidateAccessToken(token)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func Test_Unit_ValidateAccessToken_RejectsUnexpectedAlgorithm(t *testing.T) {
	setupJwtTestConfig(t)

	claims := AccessTokenClaims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-uuid-test",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	// alg=none is a classic JWT confusion attack - must be rejected outright.
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = ValidateAccessToken(signed)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func Test_Unit_GenerateRefreshToken_HashIsDeterministicAndUnlinkable(t *testing.T) {
	rawToken, hash, err := generateRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, rawToken)
	require.NotEmpty(t, hash)
	require.NotEqual(t, rawToken, hash)
	require.Equal(t, hash, hashRefreshToken(rawToken))

	_, hash2, err := generateRefreshToken()
	require.NoError(t, err)
	require.NotEqual(t, hash, hash2)
}
