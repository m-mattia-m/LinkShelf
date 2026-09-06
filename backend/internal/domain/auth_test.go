package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Unit_Auth_Login_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.refreshTokenExpiryMinutes", 60)

	hashed, err := hashPassword("correct-password")
	require.NoError(t, err)

	svc.UserRepository.
		EXPECT().
		FindByEmail("user@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Role: model.RoleUser, Password: hashed}, nil)

	svc.RefreshTokenRepository.
		EXPECT().
		Create("user-uuid-test", gomock.Any(), gomock.Any()).
		Return(nil)

	tokens, err := svc.Service.AuthService.Login("user@test.com", "correct-password")

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
}

func Test_Unit_Auth_Login_WrongPassword(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)

	hashed, err := hashPassword("correct-password")
	require.NoError(t, err)

	svc.UserRepository.
		EXPECT().
		FindByEmail("user@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Role: model.RoleUser, Password: hashed}, nil)

	tokens, err := svc.Service.AuthService.Login("user@test.com", "wrong-password")

	require.ErrorIs(t, err, ErrInvalidCredentials)
	require.Nil(t, tokens)
}

func Test_Unit_Auth_Login_NoLocalPassword(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)

	svc.UserRepository.
		EXPECT().
		FindByEmail("external-only@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Role: model.RoleUser, Password: ""}, nil)

	tokens, err := svc.Service.AuthService.Login("external-only@test.com", "anything")

	require.ErrorIs(t, err, ErrInvalidCredentials)
	require.Nil(t, tokens)
}

func Test_Unit_Auth_Login_UnknownEmail(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)

	svc.UserRepository.
		EXPECT().
		FindByEmail("nobody@test.com").
		Return(nil, nil)

	tokens, err := svc.Service.AuthService.Login("nobody@test.com", "anything")

	require.ErrorIs(t, err, ErrInvalidCredentials)
	require.Nil(t, tokens)
}

func Test_Unit_Auth_Refresh_Success_RotatesToken(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.refreshTokenExpiryMinutes", 60)

	rawToken, hash, err := generateRefreshToken()
	require.NoError(t, err)

	svc.RefreshTokenRepository.
		EXPECT().
		GetByHash(hash).
		Return(&repository.RefreshToken{Id: "rt-1", UserId: "user-uuid-test", TokenHash: hash, ExpiresAt: time.Now().Add(time.Hour)}, nil)

	svc.RefreshTokenRepository.
		EXPECT().
		DeleteByHash(hash).
		Return(nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{Role: model.RoleUser}}, nil)

	svc.RefreshTokenRepository.
		EXPECT().
		Create("user-uuid-test", gomock.Any(), gomock.Any()).
		Return(nil)

	tokens, err := svc.Service.AuthService.Refresh(rawToken)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEqual(t, rawToken, tokens.RefreshToken)
}

func Test_Unit_Auth_Refresh_Expired(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)

	rawToken, hash, err := generateRefreshToken()
	require.NoError(t, err)

	svc.RefreshTokenRepository.
		EXPECT().
		GetByHash(hash).
		Return(&repository.RefreshToken{Id: "rt-1", UserId: "user-uuid-test", TokenHash: hash, ExpiresAt: time.Now().Add(-time.Minute)}, nil)

	tokens, err := svc.Service.AuthService.Refresh(rawToken)

	require.ErrorIs(t, err, ErrInvalidToken)
	require.Nil(t, tokens)
}

func Test_Unit_Auth_Refresh_UnknownToken(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)

	rawToken, hash, err := generateRefreshToken()
	require.NoError(t, err)

	svc.RefreshTokenRepository.
		EXPECT().
		GetByHash(hash).
		Return(nil, nil)

	tokens, err := svc.Service.AuthService.Refresh(rawToken)

	require.ErrorIs(t, err, ErrInvalidToken)
	require.Nil(t, tokens)
}

func Test_Unit_Auth_Logout_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	rawToken, hash, err := generateRefreshToken()
	require.NoError(t, err)

	svc.RefreshTokenRepository.
		EXPECT().
		DeleteByHash(hash).
		Return(nil)

	err = svc.Service.AuthService.Logout(rawToken)

	require.NoError(t, err)
}

func Test_Unit_Auth_Logout_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	rawToken, hash, err := generateRefreshToken()
	require.NoError(t, err)

	svc.RefreshTokenRepository.
		EXPECT().
		DeleteByHash(hash).
		Return(errors.New("db error"))

	err = svc.Service.AuthService.Logout(rawToken)

	require.ErrorContains(t, err, "db error")
}

func Test_Unit_Auth_Oidc_NotConfigured(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	_, err := svc.Service.AuthService.OidcAuthorizationURL()
	require.ErrorIs(t, err, ErrOidcNotConfigured)

	_, err = svc.Service.AuthService.OidcCallback(context.Background(), "code", "state", nil)
	require.ErrorIs(t, err, ErrOidcNotConfigured)
}
