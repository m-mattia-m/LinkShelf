//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_API_Auth_Login_Success(t *testing.T) {
	randUuid, err := uuid.NewV7()
	require.NoError(t, err)
	email := "auth-login-" + ShortUUID(randUuid.String()) + "@test.com"

	_, err = TestService.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Email: email, FirstName: "Auth", LastName: "Login"},
		Password: "correct-password",
	}, true)
	require.NoError(t, err)

	resp := doRequest(t, http.MethodPost, "/v1/auth/login", strings.NewReader(ObjectToJSON(model.LoginRequest{
		Email: email, Password: "correct-password",
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var tokens model.TokenPair
	require.NoError(t, json.Unmarshal(body, &tokens))
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
}

func Test_API_Auth_Login_WrongPassword(t *testing.T) {
	randUuid, err := uuid.NewV7()
	require.NoError(t, err)
	email := "auth-login-wrong-" + ShortUUID(randUuid.String()) + "@test.com"

	_, err = TestService.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Email: email, FirstName: "Auth", LastName: "Login"},
		Password: "correct-password",
	}, true)
	require.NoError(t, err)

	resp := doRequest(t, http.MethodPost, "/v1/auth/login", strings.NewReader(ObjectToJSON(model.LoginRequest{
		Email: email, Password: "wrong-password",
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_API_Auth_Refresh_RotatesToken(t *testing.T) {
	randUuid, err := uuid.NewV7()
	require.NoError(t, err)
	email := "auth-refresh-" + ShortUUID(randUuid.String()) + "@test.com"
	const password = "secret"
	_, err = TestService.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Email: email, FirstName: "Auth", LastName: "Refresh"},
		Password: password,
	}, true)
	require.NoError(t, err)

	loginResp := doRequest(t, http.MethodPost, "/v1/auth/login", strings.NewReader(ObjectToJSON(model.LoginRequest{
		Email: email, Password: password,
	})))
	loginBody, err := io.ReadAll(loginResp.Body)
	require.NoError(t, err)
	_ = loginResp.Body.Close()
	var initialTokens model.TokenPair
	require.NoError(t, json.Unmarshal(loginBody, &initialTokens))

	refreshResp := doRequest(t, http.MethodPost, "/v1/auth/refresh", strings.NewReader(ObjectToJSON(model.RefreshRequest{
		RefreshToken: initialTokens.RefreshToken,
	})))
	defer func() { _ = refreshResp.Body.Close() }()
	require.Equal(t, http.StatusOK, refreshResp.StatusCode)

	refreshBody, err := io.ReadAll(refreshResp.Body)
	require.NoError(t, err)
	var rotatedTokens model.TokenPair
	require.NoError(t, json.Unmarshal(refreshBody, &rotatedTokens))
	require.NotEmpty(t, rotatedTokens.RefreshToken)
	require.NotEqual(t, initialTokens.RefreshToken, rotatedTokens.RefreshToken)

	// Single-use rotation: the original refresh token must no longer work.
	reuseResp := doRequest(t, http.MethodPost, "/v1/auth/refresh", strings.NewReader(ObjectToJSON(model.RefreshRequest{
		RefreshToken: initialTokens.RefreshToken,
	})))
	defer func() { _ = reuseResp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, reuseResp.StatusCode)
}

func Test_API_Auth_Refresh_InvalidToken(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/auth/refresh", strings.NewReader(ObjectToJSON(model.RefreshRequest{
		RefreshToken: "not-a-real-refresh-token",
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_API_Auth_Logout_InvalidatesRefreshToken(t *testing.T) {
	randUuid, err := uuid.NewV7()
	require.NoError(t, err)
	email := "auth-logout-" + ShortUUID(randUuid.String()) + "@test.com"
	const password = "secret"
	_, err = TestService.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Email: email, FirstName: "Auth", LastName: "Logout"},
		Password: password,
	}, true)
	require.NoError(t, err)

	loginResp := doRequest(t, http.MethodPost, "/v1/auth/login", strings.NewReader(ObjectToJSON(model.LoginRequest{
		Email: email, Password: password,
	})))
	loginBody, err := io.ReadAll(loginResp.Body)
	require.NoError(t, err)
	_ = loginResp.Body.Close()
	var tokens model.TokenPair
	require.NoError(t, json.Unmarshal(loginBody, &tokens))

	logoutResp := doRequest(t, http.MethodPost, "/v1/auth/logout", strings.NewReader(ObjectToJSON(model.RefreshRequest{
		RefreshToken: tokens.RefreshToken,
	})))
	defer func() { _ = logoutResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, logoutResp.StatusCode)

	refreshResp := doRequest(t, http.MethodPost, "/v1/auth/refresh", strings.NewReader(ObjectToJSON(model.RefreshRequest{
		RefreshToken: tokens.RefreshToken,
	})))
	defer func() { _ = refreshResp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, refreshResp.StatusCode)
}

func Test_API_Auth_OidcLogin_NotConfigured(t *testing.T) {
	// This test harness runs with authentication.type=LOCAL, so OIDC is
	// disabled - the endpoint must fail cleanly rather than panic on a nil
	// OIDC client.
	resp := doRequest(t, http.MethodGet, "/v1/auth/oidc/login", nil)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
