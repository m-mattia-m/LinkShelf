//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/mailer"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// capturingMailer stands in for SMTP so a test can read the reset link that
// would have been emailed.
type capturingMailer struct {
	mu       sync.Mutex
	messages []mailer.Message
}

func (m *capturingMailer) Send(msg mailer.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
	return nil
}

func (m *capturingMailer) sentTo(email string) []mailer.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []mailer.Message
	for _, msg := range m.messages {
		if msg.To == email {
			out = append(out, msg)
		}
	}
	return out
}

// passwordResetEnabled turns the feature on for one test and swaps in a
// service whose mailer is captured. The HTTP handlers look the service up on
// every request, so replacing the field is enough.
func passwordResetEnabled(t *testing.T) *capturingMailer {
	t.Helper()
	captured := &capturingMailer{}
	original := TestService.EmailVerificationService

	config.Set("authentication.passwordReset.enabled", true)
	config.Set("authentication.passwordReset.tokenExpiryMinutes", 60)
	config.Set("app.frontendUrl", "http://localhost:3000")
	TestService.EmailVerificationService = domain.NewEmailVerificationService(TestRepository, TestService, captured)

	t.Cleanup(func() {
		TestService.EmailVerificationService = original
		config.Set("authentication.passwordReset.enabled", false)
	})
	return captured
}

func newAccount(t *testing.T, password string) (email string) {
	t.Helper()
	email = "reset-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12] + "@test.com"
	_, err := TestService.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Username: uniqueUsername(), Email: email, FirstName: "Reset", LastName: "Me"},
		Password: password,
	}, true)
	require.NoError(t, err)
	return email
}

func login(t *testing.T, email, password string) (status int, tokens model.TokenPair) {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/v1/auth/login", strings.NewReader(ObjectToJSON(model.LoginRequest{Email: email, Password: password})))
	status = resp.StatusCode
	body := readBody(t, resp)
	if status == http.StatusOK {
		require.NoError(t, json.Unmarshal(body, &tokens))
	}
	return status, tokens
}

func forgotPassword(t *testing.T, email string) *http.Response {
	t.Helper()
	return doRequest(t, http.MethodPost, "/v1/auth/forgot-password", strings.NewReader(ObjectToJSON(model.ForgotPasswordRequest{Email: email})))
}

func resetPassword(t *testing.T, token, newPassword string) *http.Response {
	t.Helper()
	return doRequest(t, http.MethodPost, "/v1/auth/reset-password", strings.NewReader(ObjectToJSON(model.ResetPasswordRequest{Token: token, NewPassword: newPassword})))
}

var resetLink = regexp.MustCompile(`http://localhost:3000/auth/reset-password\?token=([A-Za-z0-9_-]+)`)

func tokenFrom(t *testing.T, msg mailer.Message) string {
	t.Helper()
	match := resetLink.FindStringSubmatch(msg.TextBody)
	require.NotNil(t, match, "the email contains a reset link: %s", msg.TextBody)
	return match[1]
}

func Test_API_PasswordReset_FullFlow(t *testing.T) {
	captured := passwordResetEnabled(t)
	email := newAccount(t, "old-password-123")

	status, oldTokens := login(t, email, "old-password-123")
	require.Equal(t, http.StatusOK, status)

	// Ask for a reset: the caller learns nothing, the email carries the link.
	resp := forgotPassword(t, email)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = readBody(t, resp)

	messages := captured.sentTo(email)
	require.Len(t, messages, 1)
	require.Contains(t, messages[0].Subject, "Reset your")
	token := tokenFrom(t, messages[0])

	// Too short a password is refused and does not burn the link.
	resp = resetPassword(t, token, "short")
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	_ = readBody(t, resp)

	resp = resetPassword(t, token, "brand-new-password-456")
	require.Equal(t, http.StatusNoContent, resp.StatusCode, "the link still works after the rejected attempt")
	_ = readBody(t, resp)

	status, _ = login(t, email, "old-password-123")
	require.Equal(t, http.StatusUnauthorized, status, "the old password stops working")
	status, _ = login(t, email, "brand-new-password-456")
	require.Equal(t, http.StatusOK, status)

	// Everyone signed in with the old password is signed out.
	refresh := doRequest(t, http.MethodPost, "/v1/auth/refresh", strings.NewReader(ObjectToJSON(map[string]string{"refresh_token": oldTokens.RefreshToken})))
	require.Equal(t, http.StatusUnauthorized, refresh.StatusCode)
	_ = readBody(t, refresh)

	// A link works once.
	resp = resetPassword(t, token, "another-new-password-789")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = readBody(t, resp)
}

func Test_API_PasswordReset_UnknownAddressLooksIdentical(t *testing.T) {
	captured := passwordResetEnabled(t)
	known := newAccount(t, "old-password-123")
	unknown := "nobody-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12] + "@test.com"

	knownResp := forgotPassword(t, known)
	knownBody := readBody(t, knownResp)
	unknownResp := forgotPassword(t, unknown)
	unknownBody := readBody(t, unknownResp)

	require.Equal(t, knownResp.StatusCode, unknownResp.StatusCode)
	require.Equal(t, string(knownBody), string(unknownBody))
	require.Len(t, captured.sentTo(known), 1)
	require.Empty(t, captured.sentTo(unknown), "no email for an address without an account")
}

func Test_API_PasswordReset_RepeatedRequestsSendOneEmail(t *testing.T) {
	captured := passwordResetEnabled(t)
	email := newAccount(t, "old-password-123")

	for i := 0; i < 3; i++ {
		resp := forgotPassword(t, email)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		_ = readBody(t, resp)
	}

	require.Len(t, captured.sentTo(email), 1, "requests inside the cooldown do not send more mail")
}

func Test_API_PasswordReset_InvalidTokenIsRejected(t *testing.T) {
	passwordResetEnabled(t)

	resp := resetPassword(t, "not-a-real-token", "brand-new-password-456")

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = readBody(t, resp)
}

func Test_API_PasswordReset_OtherEmailLinksCannotResetAPassword(t *testing.T) {
	captured := passwordResetEnabled(t)
	email := newAccount(t, "old-password-123")
	resp := forgotPassword(t, email)
	_ = readBody(t, resp)
	token := tokenFrom(t, captured.sentTo(email)[0])

	// The same token presented to the invite endpoint is not accepted there.
	setResp := doRequest(t, http.MethodPost, "/v1/auth/set-password", strings.NewReader(ObjectToJSON(model.SetPasswordRequest{Token: token, NewPassword: "sneaky-password-123"})))
	require.Equal(t, http.StatusBadRequest, setResp.StatusCode)
	_ = readBody(t, setResp)

	status, _ := login(t, email, "old-password-123")
	require.Equal(t, http.StatusOK, status, "the password is unchanged")
}

func Test_API_PasswordReset_Disabled(t *testing.T) {
	// Feature switched off (the test config default).
	config.Set("authentication.passwordReset.enabled", false)

	resp := forgotPassword(t, "someone@test.com")
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	_ = readBody(t, resp)

	resp = resetPassword(t, "any-token", "brand-new-password-456")
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	_ = readBody(t, resp)
}

func Test_API_PasswordReset_SettingsExposeTheFlag(t *testing.T) {
	read := func() bool {
		resp := doRequest(t, http.MethodGet, "/v1/settings?language_code=en", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var settings model.SettingPageBody
		require.NoError(t, json.Unmarshal(readBody(t, resp), &settings))
		return settings.PasswordResetEnabled
	}

	require.False(t, read())

	passwordResetEnabled(t)
	require.True(t, read())
}
