package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_API_Login_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		Login("user@test.com", "correct-password").
		Return(&model.TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token"}, nil)

	handler := Login(svc.Service)
	resp, err := handler(context.Background(), &model.LoginRequestBody{
		Body: model.LoginRequest{Email: "user@test.com", Password: "correct-password"},
	})

	require.NoError(t, err)
	require.Equal(t, "access-token", resp.Body.AccessToken)
	require.Equal(t, "refresh-token", resp.Body.RefreshToken)
}

func Test_API_Login_InvalidCredentials(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		Login("user@test.com", "wrong-password").
		Return(nil, domain.ErrInvalidCredentials)

	handler := Login(svc.Service)
	_, err := handler(context.Background(), &model.LoginRequestBody{
		Body: model.LoginRequest{Email: "user@test.com", Password: "wrong-password"},
	})

	require.Error(t, err)
}

func Test_API_Refresh_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		Refresh("raw-refresh-token").
		Return(&model.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh"}, nil)

	handler := Refresh(svc.Service)
	resp, err := handler(context.Background(), &model.RefreshRequestBody{
		Body: model.RefreshRequest{RefreshToken: "raw-refresh-token"},
	})

	require.NoError(t, err)
	require.Equal(t, "new-access", resp.Body.AccessToken)
}

func Test_API_Refresh_InvalidToken(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		Refresh("bad-token").
		Return(nil, domain.ErrInvalidToken)

	handler := Refresh(svc.Service)
	_, err := handler(context.Background(), &model.RefreshRequestBody{
		Body: model.RefreshRequest{RefreshToken: "bad-token"},
	})

	require.Error(t, err)
}

func Test_API_Logout_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		Logout("raw-refresh-token").
		Return(nil)

	handler := Logout(svc.Service)
	_, err := handler(context.Background(), &model.LogoutRequestBody{
		Body: model.RefreshRequest{RefreshToken: "raw-refresh-token"},
	})

	require.NoError(t, err)
}

func Test_API_Logout_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		Logout("raw-refresh-token").
		Return(errors.New("db unavailable"))

	handler := Logout(svc.Service)
	_, err := handler(context.Background(), &model.LogoutRequestBody{
		Body: model.RefreshRequest{RefreshToken: "raw-refresh-token"},
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to logout")
}

func Test_API_OidcLogin_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		OidcAuthorizationURL().
		Return(&model.OidcLoginResponseBody{AuthorizationUrl: "https://idp.example.com/authorize", State: "state-value"}, nil)

	handler := OidcLogin(svc.Service)
	resp, err := handler(context.Background(), &struct{}{})

	require.NoError(t, err)
	require.Equal(t, "https://idp.example.com/authorize", resp.Body.AuthorizationUrl)
}

func Test_API_OidcLogin_NotConfigured(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		OidcAuthorizationURL().
		Return(nil, domain.ErrOidcNotConfigured)

	handler := OidcLogin(svc.Service)
	_, err := handler(context.Background(), &struct{}{})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to start OIDC login")
}

func Test_API_OidcCallback_AnonymousLogin_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		OidcCallback(gomock.Any(), "auth-code", "state-value", (*string)(nil)).
		Return(&model.TokenPair{AccessToken: "access", RefreshToken: "refresh"}, nil)

	handler := OidcCallback(svc.Service)
	resp, err := handler(context.Background(), &model.OidcCallbackRequestBody{
		Body: model.OidcCallbackRequest{Code: "auth-code", State: "state-value"},
	})

	require.NoError(t, err)
	require.Equal(t, "access", resp.Body.AccessToken)
}

func Test_API_OidcCallback_InvalidBearer_TreatedAsAnonymous(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		OidcCallback(gomock.Any(), "auth-code", "state-value", (*string)(nil)).
		Return(&model.TokenPair{AccessToken: "access", RefreshToken: "refresh"}, nil)

	handler := OidcCallback(svc.Service)
	_, err := handler(context.Background(), &model.OidcCallbackRequestBody{
		Authorization: "Bearer not-a-real-token",
		Body:          model.OidcCallbackRequest{Code: "auth-code", State: "state-value"},
	})

	require.NoError(t, err)
}

func Test_API_OidcCallback_LinkMode_WithValidBearer(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	config.Reset()
	config.Set("authentication.jwtSecret", "test-secret")

	token := signTestAccessToken(t, "current-user-id", model.RoleUser)

	svc.AuthService.
		EXPECT().
		OidcCallback(gomock.Any(), "auth-code", "state-value", gomock.Not(gomock.Nil())).
		DoAndReturn(func(_ context.Context, code, state string, currentUserId *string) (*model.TokenPair, error) {
			require.NotNil(t, currentUserId)
			require.Equal(t, "current-user-id", *currentUserId)
			return &model.TokenPair{AccessToken: "linked-access", RefreshToken: "linked-refresh"}, nil
		})

	handler := OidcCallback(svc.Service)
	resp, err := handler(context.Background(), &model.OidcCallbackRequestBody{
		Authorization: "Bearer " + token,
		Body:          model.OidcCallbackRequest{Code: "auth-code", State: "state-value"},
	})

	require.NoError(t, err)
	require.Equal(t, "linked-access", resp.Body.AccessToken)
}

func Test_API_OidcCallback_EmailNotVerified_MapsToConflict(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		OidcCallback(gomock.Any(), "auth-code", "state-value", (*string)(nil)).
		Return(nil, domain.ErrEmailNotVerified)

	handler := OidcCallback(svc.Service)
	_, err := handler(context.Background(), &model.OidcCallbackRequestBody{
		Body: model.OidcCallbackRequest{Code: "auth-code", State: "state-value"},
	})

	require.Error(t, err)
}

func Test_API_OidcCallback_EmailNotVerifiedForLinking_MapsToConflict(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		OidcCallback(gomock.Any(), "auth-code", "state-value", (*string)(nil)).
		Return(nil, domain.ErrEmailNotVerifiedForLinking)

	handler := OidcCallback(svc.Service)
	_, err := handler(context.Background(), &model.OidcCallbackRequestBody{
		Body: model.OidcCallbackRequest{Code: "auth-code", State: "state-value"},
	})

	require.Error(t, err)
}

func Test_API_OidcCallback_GenericError(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.AuthService.
		EXPECT().
		OidcCallback(gomock.Any(), "auth-code", "state-value", (*string)(nil)).
		Return(nil, errors.New("exchange failed"))

	handler := OidcCallback(svc.Service)
	_, err := handler(context.Background(), &model.OidcCallbackRequestBody{
		Body: model.OidcCallbackRequest{Code: "auth-code", State: "state-value"},
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to complete OIDC login")
}

// signTestAccessToken mints a token in the same shape/signing scheme as
// domain's own (unexported) issueAccessToken, so controller tests can
// exercise the Authorization-header parsing path without needing a real
// login first.
func signTestAccessToken(t *testing.T, userId, role string) string {
	t.Helper()
	claims := domain.AccessTokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.String("authentication.jwtSecret")))
	require.NoError(t, err)
	return signed
}
