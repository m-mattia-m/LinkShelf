package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/oidcclient"
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

func Test_Unit_Auth_Login_Success_VerificationEnabledAndVerified(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.refreshTokenExpiryMinutes", 60)
	config.Set("authentication.emailVerification.enabled", true)

	hashed, err := hashPassword("correct-password")
	require.NoError(t, err)

	svc.UserRepository.
		EXPECT().
		FindByEmail("user@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Role: model.RoleUser, Password: hashed, EmailVerified: true}, nil)

	svc.RefreshTokenRepository.
		EXPECT().
		Create("user-uuid-test", gomock.Any(), gomock.Any()).
		Return(nil)

	tokens, err := svc.Service.AuthService.Login("user@test.com", "correct-password")

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
}

func Test_Unit_Auth_Login_Pending_VerificationEnabledButNotVerified(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.emailVerification.enabled", true)
	config.Set("authentication.emailVerification.tokenExpiryHours", 24)

	hashed, err := hashPassword("correct-password")
	require.NoError(t, err)

	svc.UserRepository.
		EXPECT().
		FindByEmail("user@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Email: "user@test.com", Role: model.RoleUser, Password: hashed, EmailVerified: false}, nil).
		Times(2) // once by Login, once by the auto-triggered Resend

	svc.EmailActionTokenRepository.
		EXPECT().
		GetLatestByUserIdAndAction("user-uuid-test", repository.EmailActionVerify).
		Return(nil, nil)
	svc.EmailActionTokenRepository.
		EXPECT().
		DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionVerify).
		Return(nil)
	svc.EmailActionTokenRepository.
		EXPECT().
		Create("user-uuid-test", gomock.Any(), repository.EmailActionVerify, gomock.Any()).
		Return(nil)
	svc.Mailer.EXPECT().Send(gomock.Any()).Return(nil)

	tokens, err := svc.Service.AuthService.Login("user@test.com", "correct-password")

	require.ErrorIs(t, err, ErrEmailVerificationPending)
	require.Nil(t, tokens)
}

func Test_Unit_Auth_Login_Pending_InvitedAccountNoPassword(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.emailVerification.enabled", true)
	config.Set("authentication.emailVerification.tokenExpiryHours", 24)

	svc.UserRepository.
		EXPECT().
		FindByEmail("invited@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Email: "invited@test.com", Role: model.RoleUser, Password: ""}, nil).
		Times(2)

	svc.EmailActionTokenRepository.
		EXPECT().
		GetLatestByUserIdAndAction("user-uuid-test", repository.EmailActionSetPassword).
		Return(nil, nil)
	svc.EmailActionTokenRepository.
		EXPECT().
		DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionSetPassword).
		Return(nil)
	svc.EmailActionTokenRepository.
		EXPECT().
		Create("user-uuid-test", gomock.Any(), repository.EmailActionSetPassword, gomock.Any()).
		Return(nil)
	svc.Mailer.EXPECT().Send(gomock.Any()).Return(nil)

	tokens, err := svc.Service.AuthService.Login("invited@test.com", "anything")

	require.ErrorIs(t, err, ErrEmailVerificationPending)
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

	require.False(t, svc.Service.AuthService.IsOidcEnabled())
}

func Test_Unit_Auth_IsOidcEnabled_True(t *testing.T) {
	repo := &repository.Repository{}
	service := NewService(repo, &oidcclient.Client{}, nil)

	require.True(t, service.AuthService.IsOidcEnabled())
}

func Test_Unit_Auth_ResolveOidcIdentity_LinkMode_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.refreshTokenExpiryMinutes", 60)

	currentUserId := "current-user-id"
	identity := &oidcclient.Identity{Subject: "provider-subject"}

	svc.UserRepository.
		EXPECT().
		LinkProvider(currentUserId, model.ProviderOIDC, identity.Subject).
		Return(nil)

	svc.UserRepository.
		EXPECT().
		Get(currentUserId).
		Return(&model.User{Id: currentUserId, UserBase: model.UserBase{Role: model.RoleUser}}, nil)

	svc.RefreshTokenRepository.
		EXPECT().
		Create(currentUserId, gomock.Any(), gomock.Any()).
		Return(nil)

	authSvc := svc.Service.AuthService.(*authServiceImpl)
	tokens, err := authSvc.resolveOidcIdentity(identity, &currentUserId)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
}

func Test_Unit_Auth_ResolveOidcIdentity_ReturningUser_MatchesByProviderId(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.refreshTokenExpiryMinutes", 60)

	identity := &oidcclient.Identity{Subject: "provider-subject", Email: "user@test.com", EmailVerified: true}

	svc.UserRepository.
		EXPECT().
		FindByProviderId(identity.Subject).
		Return(&repository.AuthRecord{Id: "existing-user-id", Role: model.RoleUser}, nil)

	svc.RefreshTokenRepository.
		EXPECT().
		Create("existing-user-id", gomock.Any(), gomock.Any()).
		Return(nil)

	authSvc := svc.Service.AuthService.(*authServiceImpl)
	tokens, err := authSvc.resolveOidcIdentity(identity, nil)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
}

func Test_Unit_Auth_ResolveOidcIdentity_AutoLink_VerifiedEmail_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.refreshTokenExpiryMinutes", 60)

	identity := &oidcclient.Identity{Subject: "provider-subject", Email: "user@test.com", EmailVerified: true}

	svc.UserRepository.
		EXPECT().
		FindByProviderId(identity.Subject).
		Return(nil, nil)

	svc.UserRepository.
		EXPECT().
		FindByEmail(identity.Email).
		Return(&repository.AuthRecord{Id: "existing-user-id", Role: model.RoleUser}, nil)

	svc.UserRepository.
		EXPECT().
		LinkProvider("existing-user-id", model.ProviderOIDC, identity.Subject).
		Return(nil)

	svc.RefreshTokenRepository.
		EXPECT().
		Create("existing-user-id", gomock.Any(), gomock.Any()).
		Return(nil)

	authSvc := svc.Service.AuthService.(*authServiceImpl)
	tokens, err := authSvc.resolveOidcIdentity(identity, nil)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
}

// Test_Unit_Auth_ResolveOidcIdentity_AutoLink_UnverifiedEmail_Fails covers the
// exact scenario reported as confusing: an account with this email DOES
// exist, so the "already exists" wording is accurate here - see the sibling
// AutoProvision_UnverifiedEmail_Fails test for the case where it isn't.
func Test_Unit_Auth_ResolveOidcIdentity_AutoLink_UnverifiedEmail_Fails(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	identity := &oidcclient.Identity{Subject: "provider-subject", Email: "user@test.com", EmailVerified: false}

	svc.UserRepository.
		EXPECT().
		FindByProviderId(identity.Subject).
		Return(nil, nil)

	svc.UserRepository.
		EXPECT().
		FindByEmail(identity.Email).
		Return(&repository.AuthRecord{Id: "existing-user-id", Role: model.RoleUser}, nil)

	authSvc := svc.Service.AuthService.(*authServiceImpl)
	tokens, err := authSvc.resolveOidcIdentity(identity, nil)

	require.ErrorIs(t, err, ErrEmailNotVerifiedForLinking)
	require.Nil(t, tokens)
}

func Test_Unit_Auth_ResolveOidcIdentity_AutoProvision_VerifiedEmail_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupJwtTestConfig(t)
	config.Set("authentication.refreshTokenExpiryMinutes", 60)

	identity := &oidcclient.Identity{
		Subject: "provider-subject", Email: "new-user@test.com", EmailVerified: true,
		FirstName: "First", LastName: "Last",
	}

	svc.UserRepository.
		EXPECT().
		FindByProviderId(identity.Subject).
		Return(nil, nil)

	svc.UserRepository.
		EXPECT().
		FindByEmail(identity.Email).
		Return(nil, nil)

	svc.UserRepository.
		EXPECT().
		CreateExternal(identity.Email, identity.FirstName, identity.LastName, model.ProviderOIDC, identity.Subject).
		Return("new-user-id", nil)

	svc.RefreshTokenRepository.
		EXPECT().
		Create("new-user-id", gomock.Any(), gomock.Any()).
		Return(nil)

	authSvc := svc.Service.AuthService.(*authServiceImpl)
	tokens, err := authSvc.resolveOidcIdentity(identity, nil)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
}

// Test_Unit_Auth_ResolveOidcIdentity_AutoProvision_UnverifiedEmail_Fails
// covers the case that was reported as confusing: no account with this email
// exists at all (e.g. it was just deleted), yet the provider's unverified
// email still refuses the login. ErrEmailNotVerified (not
// ErrEmailNotVerifiedForLinking) must be returned here so the caller isn't
// told an account exists when it doesn't, and CreateExternal must never be
// called for an unverified identity.
func Test_Unit_Auth_ResolveOidcIdentity_AutoProvision_UnverifiedEmail_Fails(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	identity := &oidcclient.Identity{Subject: "provider-subject", Email: "nobody@test.com", EmailVerified: false}

	svc.UserRepository.
		EXPECT().
		FindByProviderId(identity.Subject).
		Return(nil, nil)

	svc.UserRepository.
		EXPECT().
		FindByEmail(identity.Email).
		Return(nil, nil)

	authSvc := svc.Service.AuthService.(*authServiceImpl)
	tokens, err := authSvc.resolveOidcIdentity(identity, nil)

	require.ErrorIs(t, err, ErrEmailNotVerified)
	require.NotErrorIs(t, err, ErrEmailNotVerifiedForLinking)
	require.Nil(t, tokens)
}
