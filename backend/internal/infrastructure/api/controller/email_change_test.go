package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func requireStatus(t *testing.T, err error, status int) {
	t.Helper()
	var statusErr huma.StatusError
	require.True(t, errors.As(err, &statusErr), "expected a huma.StatusError, got %v", err)
	require.Equal(t, status, statusErr.GetStatus())
}

func Test_API_ConfirmEmailChange_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.EmailVerificationService.EXPECT().ConfirmEmailChange("raw-token").Return(nil)

	_, err := ConfirmEmailChange(svc.Service)(context.Background(), &model.VerifyEmailRequestBody{
		Body: model.VerifyEmailRequest{Token: "raw-token"},
	})

	require.NoError(t, err)
}

func Test_API_ConfirmEmailChange_MapsErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		err    error
		status int
	}{
		"invalid token":  {domain.ErrInvalidToken, http.StatusBadRequest},
		"address taken":  {domain.ErrConflict, http.StatusConflict},
		"internal error": {errors.New("db down"), http.StatusBadRequest},
	} {
		t.Run(name, func(t *testing.T) {
			svc := NewMockDomainService(t)
			defer svc.Ctrl.Finish()

			svc.EmailVerificationService.EXPECT().ConfirmEmailChange("raw-token").Return(tc.err)

			_, err := ConfirmEmailChange(svc.Service)(context.Background(), &model.VerifyEmailRequestBody{
				Body: model.VerifyEmailRequest{Token: "raw-token"},
			})

			requireStatus(t, err, tc.status)
		})
	}
}

func Test_API_UpdateUser_EmailChangeRateLimited_Is429(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.UserService.EXPECT().Update("user-uuid-test", gomock.Any(), false).Return(nil, domain.ErrTooManyRequests)

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	_, err := UpdateUser(svc.Service)(ctx, &model.UserFilterFilterAndBody{
		UserRequestFilter: model.UserRequestFilter{UserId: "user-uuid-test"},
	})

	requireStatus(t, err, http.StatusTooManyRequests)
}

func Test_API_UpdateUser_EmailTaken_Is409(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.UserService.EXPECT().Update("user-uuid-test", gomock.Any(), false).Return(nil, domain.ErrConflict)

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	_, err := UpdateUser(svc.Service)(ctx, &model.UserFilterFilterAndBody{
		UserRequestFilter: model.UserRequestFilter{UserId: "user-uuid-test"},
	})

	requireStatus(t, err, http.StatusConflict)
}

func Test_API_Login_LocalAuthDisabled_Is403(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.EXPECT().Login("user@test.com", "password").Return(nil, domain.ErrLocalAuthDisabled)

	_, err := Login(svc.Service)(context.Background(), &model.LoginRequestBody{
		Body: model.LoginRequest{Email: "user@test.com", Password: "password"},
	})

	requireStatus(t, err, http.StatusForbidden)
}

func Test_API_OidcCallback_AccountNotLinkable_Is409(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		OidcCallback(gomock.Any(), "auth-code", "state-value", (*string)(nil)).
		Return(nil, domain.ErrAccountNotLinkable)

	_, err := OidcCallback(svc.Service)(context.Background(), &model.OidcCallbackRequestBody{
		Body: model.OidcCallbackRequest{Code: "auth-code", State: "state-value"},
	})

	requireStatus(t, err, http.StatusConflict)
	require.ErrorContains(t, err, "can't be linked")
}

// Without local auth, sign-up and password reset are reported off.
func Test_SettingPageFlags_LocalAuthDisabledTurnsOffSignUpAndPasswordReset(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("authentication.registrationEnabled", true)
	config.Set("authentication.passwordReset.enabled", true)
	config.Set("authentication.localAuthEnabled", false)

	svc.AuthService.EXPECT().IsOidcEnabled().Return(true)

	flags := settingPageFlags(svc.Service)

	require.False(t, flags.LocalAuthEnabled)
	require.False(t, flags.RegistrationEnabled)
	require.False(t, flags.PasswordResetEnabled)
	require.True(t, flags.OidcEnabled)
}

func Test_SettingPageFlags_LocalAuthOnByDefault(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("authentication.registrationEnabled", true)
	config.Set("authentication.passwordReset.enabled", true)

	svc.AuthService.EXPECT().IsOidcEnabled().Return(false)

	flags := settingPageFlags(svc.Service)

	require.True(t, flags.LocalAuthEnabled)
	require.True(t, flags.RegistrationEnabled)
	require.True(t, flags.PasswordResetEnabled)
}
