//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"fmt"
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

func Test_API_Auth_Login_Blocked_WhenEmailVerificationEnabledAndNotVerified(t *testing.T) {
	config.Set("authentication.emailVerification.enabled", true)
	defer config.Set("authentication.emailVerification.enabled", false)

	randUuid, err := uuid.NewV7()
	require.NoError(t, err)
	email := "auth-login-pending-" + ShortUUID(randUuid.String()) + "@test.com"

	created, err := TestService.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Email: email, FirstName: "Auth", LastName: "Pending"},
		Password: "correct-password",
	}, true)
	require.NoError(t, err)
	require.False(t, created.EmailVerified)

	resp := doRequest(t, http.MethodPost, "/v1/auth/login", strings.NewReader(ObjectToJSON(model.LoginRequest{
		Email: email, Password: "correct-password",
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func Test_API_Auth_Login_Succeeds_AfterAdminMarksVerified(t *testing.T) {
	// Create the admin (and log it in) before flipping the gate on - an
	// ordinary admin account is not exempt from verification either (only
	// the config-driven bootstrap admin is), so this must happen first.
	_, adminToken := createTestAdmin(t)

	config.Set("authentication.emailVerification.enabled", true)
	defer config.Set("authentication.emailVerification.enabled", false)

	randUuid, err := uuid.NewV7()
	require.NoError(t, err)
	email := "auth-login-verified-" + ShortUUID(randUuid.String()) + "@test.com"

	created, err := TestService.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Email: email, FirstName: "Auth", LastName: "Verified"},
		Password: "correct-password",
	}, true)
	require.NoError(t, err)

	verifyResp := doAuthedRequest(t, http.MethodPatch, fmt.Sprintf("/v1/users/%s/verify", created.Id), nil, adminToken)
	defer func() { _ = verifyResp.Body.Close() }()
	require.Equal(t, http.StatusOK, verifyResp.StatusCode)

	resp := doRequest(t, http.MethodPost, "/v1/auth/login", strings.NewReader(ObjectToJSON(model.LoginRequest{
		Email: email, Password: "correct-password",
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func Test_API_Auth_ResendVerification_AlwaysNoContentRegardlessOfEmail(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/auth/resend-verification", strings.NewReader(ObjectToJSON(model.ResendVerificationRequest{
		Email: "no-such-account@test.com",
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func Test_API_Auth_VerifyEmail_InvalidToken(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/auth/verify-email", strings.NewReader(ObjectToJSON(model.VerifyEmailRequest{
		Token: "not-a-real-token",
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func Test_API_Auth_SetPassword_InvalidToken(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/auth/set-password", strings.NewReader(ObjectToJSON(model.SetPasswordRequest{
		Token:       "not-a-real-token",
		NewPassword: "new-secret-password",
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func Test_API_Auth_OidcLogin_NotConfigured(t *testing.T) {
	// This test harness runs with authentication.type=LOCAL, so OIDC is
	// disabled - the endpoint must fail cleanly rather than panic on a nil
	// OIDC client.
	resp := doRequest(t, http.MethodGet, "/v1/auth/oidc/login", nil)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
