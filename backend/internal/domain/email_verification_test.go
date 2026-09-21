package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/mailer"
	"backend/internal/infrastructure/repository"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupEmailVerificationTestConfig(t *testing.T) {
	t.Helper()
	config.Reset()
	config.Set("authentication.emailVerification.tokenExpiryHours", 24)
	config.Set("app.frontendUrl", "http://localhost:3000")
	config.Set("app.name", "LinkShelfTest")
}

func Test_Unit_EmailVerification_SendInitial_HasPassword_SendsVerifyLink(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailVerificationTestConfig(t)

	user := &model.User{Id: "user-uuid-test", UserBase: model.UserBase{Email: "test@test.com"}, HasPassword: true}

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

	svc.Mailer.
		EXPECT().
		Send(gomock.Any()).
		DoAndReturn(func(msg mailer.Message) error {
			require.Equal(t, "test@test.com", msg.To)
			require.Contains(t, msg.TextBody, "/auth/verify-email?token=")
			return nil
		})

	err := svc.Service.EmailVerificationService.SendInitial(user)

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_SendInitial_NoPassword_SendsSetPasswordLink(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailVerificationTestConfig(t)

	user := &model.User{Id: "user-uuid-test", UserBase: model.UserBase{Email: "invited@test.com"}, HasPassword: false}

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

	svc.Mailer.
		EXPECT().
		Send(gomock.Any()).
		DoAndReturn(func(msg mailer.Message) error {
			require.Equal(t, "invited@test.com", msg.To)
			require.Contains(t, msg.TextBody, "/auth/set-password?token=")
			return nil
		})

	err := svc.Service.EmailVerificationService.SendInitial(user)

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_Send_WithinCooldown_SkipsSending(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailVerificationTestConfig(t)

	user := &model.User{Id: "user-uuid-test", UserBase: model.UserBase{Email: "test@test.com"}, HasPassword: true}

	svc.EmailActionTokenRepository.
		EXPECT().
		GetLatestByUserIdAndAction("user-uuid-test", repository.EmailActionVerify).
		Return(&repository.EmailActionToken{CreatedAt: time.Now()}, nil)

	// No Create/Mailer.Send expectations - gomock will fail the test if
	// either is called, proving the cooldown actually short-circuits.
	err := svc.Service.EmailVerificationService.SendInitial(user)

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_Resend_UnknownEmail_NoOp(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailVerificationTestConfig(t)

	svc.UserRepository.
		EXPECT().
		FindByEmail("missing@test.com").
		Return(nil, nil)

	err := svc.Service.EmailVerificationService.Resend("missing@test.com")

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_Resend_AlreadyVerified_NoOp(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailVerificationTestConfig(t)

	svc.UserRepository.
		EXPECT().
		FindByEmail("test@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Email: "test@test.com", EmailVerified: true}, nil)

	err := svc.Service.EmailVerificationService.Resend("test@test.com")

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_Resend_PendingSetPassword_SendsSetPasswordLink(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailVerificationTestConfig(t)

	svc.UserRepository.
		EXPECT().
		FindByEmail("invited@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Email: "invited@test.com", Password: ""}, nil)

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

	err := svc.Service.EmailVerificationService.Resend("invited@test.com")

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_VerifyEmail_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.EmailActionTokenRepository.
		EXPECT().
		GetByHash(gomock.Any()).
		Return(&repository.EmailActionToken{
			UserId:    "user-uuid-test",
			Action:    repository.EmailActionVerify,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)

	svc.EmailActionTokenRepository.
		EXPECT().
		DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionVerify).
		Return(nil)

	svc.UserRepository.
		EXPECT().
		MarkVerified("user-uuid-test").
		Return(nil)

	err := svc.Service.EmailVerificationService.VerifyEmail("raw-token")

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_VerifyEmail_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.EmailActionTokenRepository.
		EXPECT().
		GetByHash(gomock.Any()).
		Return(nil, nil)

	err := svc.Service.EmailVerificationService.VerifyEmail("raw-token")

	require.ErrorIs(t, err, ErrInvalidToken)
}

func Test_Unit_EmailVerification_VerifyEmail_Expired(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.EmailActionTokenRepository.
		EXPECT().
		GetByHash(gomock.Any()).
		Return(&repository.EmailActionToken{
			UserId:    "user-uuid-test",
			Action:    repository.EmailActionVerify,
			ExpiresAt: time.Now().Add(-time.Hour),
		}, nil)

	err := svc.Service.EmailVerificationService.VerifyEmail("raw-token")

	require.ErrorIs(t, err, ErrInvalidToken)
}

func Test_Unit_EmailVerification_VerifyEmail_WrongAction(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.EmailActionTokenRepository.
		EXPECT().
		GetByHash(gomock.Any()).
		Return(&repository.EmailActionToken{
			UserId:    "user-uuid-test",
			Action:    repository.EmailActionSetPassword,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)

	err := svc.Service.EmailVerificationService.VerifyEmail("raw-token")

	require.ErrorIs(t, err, ErrInvalidToken)
}

func Test_Unit_EmailVerification_SetPassword_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.EmailActionTokenRepository.
		EXPECT().
		GetByHash(gomock.Any()).
		Return(&repository.EmailActionToken{
			UserId:    "user-uuid-test",
			Action:    repository.EmailActionSetPassword,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)

	svc.EmailActionTokenRepository.
		EXPECT().
		DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionSetPassword).
		Return(nil)

	svc.UserRepository.
		EXPECT().
		SetPassword("user-uuid-test", gomock.Any()).
		Return(nil)

	err := svc.Service.EmailVerificationService.SetPassword("raw-token", "new-secret-password")

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_MarkVerified_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		MarkVerified("user-uuid-test").
		Return(nil)

	err := svc.Service.EmailVerificationService.MarkVerified("user-uuid-test")

	require.NoError(t, err)
}

func Test_Unit_EmailVerification_Send_NilMailer_NoOp(t *testing.T) {
	repo := &repository.Repository{}
	service := NewService(repo, nil, nil)

	err := service.EmailVerificationService.SendInitial(&model.User{Id: "user-uuid-test", HasPassword: true})

	require.NoError(t, err)
}

func setupPasswordResetTestConfig(t *testing.T, enabled bool) {
	t.Helper()
	setupEmailVerificationTestConfig(t)
	config.Set("authentication.passwordReset.enabled", enabled)
	config.Set("authentication.passwordReset.tokenExpiryMinutes", 45)
}

func Test_Unit_PasswordReset_Request_Disabled(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, false)

	// No repository expectations: nothing may be looked up or sent.
	err := svc.Service.EmailVerificationService.RequestPasswordReset("test@test.com")

	require.ErrorIs(t, err, ErrPasswordResetDisabled)
}

func Test_Unit_PasswordReset_Request_UnknownEmail_LooksLikeSuccess(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, true)

	svc.UserRepository.EXPECT().FindByEmail("missing@test.com").Return(nil, nil)

	err := svc.Service.EmailVerificationService.RequestPasswordReset("missing@test.com")

	require.NoError(t, err)
}

func Test_Unit_PasswordReset_Request_SendsALinkThatExpiresWhenConfigured(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, true)

	svc.UserRepository.EXPECT().
		FindByEmail("test@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Email: "test@test.com", Password: "existing-hash", EmailVerified: true}, nil)
	svc.EmailActionTokenRepository.EXPECT().
		GetLatestByUserIdAndAction("user-uuid-test", repository.EmailActionResetPassword).
		Return(nil, nil)
	svc.EmailActionTokenRepository.EXPECT().
		DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionResetPassword).
		Return(nil)

	var storedHash string
	var storedExpiry time.Time
	svc.EmailActionTokenRepository.EXPECT().
		Create("user-uuid-test", gomock.Any(), repository.EmailActionResetPassword, gomock.Any()).
		DoAndReturn(func(_, hash, _ string, expiresAt time.Time) error {
			storedHash, storedExpiry = hash, expiresAt
			return nil
		})

	var sent mailer.Message
	svc.Mailer.EXPECT().Send(gomock.Any()).DoAndReturn(func(m mailer.Message) error {
		sent = m
		return nil
	})

	err := svc.Service.EmailVerificationService.RequestPasswordReset("test@test.com")

	require.NoError(t, err)
	require.Equal(t, "test@test.com", sent.To)
	require.Contains(t, sent.Subject, "Reset your LinkShelfTest password")
	require.Contains(t, sent.TextBody, "http://localhost:3000/auth/reset-password?token=")
	require.Contains(t, sent.HTMLBody, "http://localhost:3000/auth/reset-password?token=")
	require.Contains(t, sent.TextBody, "45 minutes")
	require.WithinDuration(t, time.Now().Add(45*time.Minute), storedExpiry, 5*time.Second)

	// Only the hash is stored, never the raw token from the link.
	rawToken := regexp.MustCompile(`token=([A-Za-z0-9_-]+)`).FindStringSubmatch(sent.TextBody)[1]
	require.Equal(t, hashOpaqueToken(rawToken), storedHash)
	require.NotContains(t, storedHash, rawToken)
}

func Test_Unit_PasswordReset_Request_AlsoWorksForAnInvitedAccountWithoutAPassword(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, true)

	svc.UserRepository.EXPECT().
		FindByEmail("invited@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Email: "invited@test.com", Password: ""}, nil)
	svc.EmailActionTokenRepository.EXPECT().GetLatestByUserIdAndAction("user-uuid-test", repository.EmailActionResetPassword).Return(nil, nil)
	svc.EmailActionTokenRepository.EXPECT().DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionResetPassword).Return(nil)
	svc.EmailActionTokenRepository.EXPECT().Create("user-uuid-test", gomock.Any(), repository.EmailActionResetPassword, gomock.Any()).Return(nil)
	svc.Mailer.EXPECT().Send(gomock.Any()).Return(nil)

	require.NoError(t, svc.Service.EmailVerificationService.RequestPasswordReset("invited@test.com"))
}

func Test_Unit_PasswordReset_Request_IsRateLimited(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, true)

	svc.UserRepository.EXPECT().
		FindByEmail("test@test.com").
		Return(&repository.AuthRecord{Id: "user-uuid-test", Email: "test@test.com", Password: "hash"}, nil)
	// A link went out a few seconds ago: no new token, no second email.
	svc.EmailActionTokenRepository.EXPECT().
		GetLatestByUserIdAndAction("user-uuid-test", repository.EmailActionResetPassword).
		Return(&repository.EmailActionToken{CreatedAt: time.Now().Add(-5 * time.Second)}, nil)

	err := svc.Service.EmailVerificationService.RequestPasswordReset("test@test.com")

	require.NoError(t, err, "a throttled request must look like any other")
}

func Test_Unit_PasswordReset_Request_PropagatesInternalErrors(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, true)

	svc.UserRepository.EXPECT().FindByEmail("test@test.com").Return(nil, errors.New("db unavailable"))

	err := svc.Service.EmailVerificationService.RequestPasswordReset("test@test.com")

	require.ErrorContains(t, err, "db unavailable")
}

func Test_Unit_PasswordReset_Reset_SetsThePasswordAndEndsEverySession(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, true)

	svc.EmailActionTokenRepository.EXPECT().
		GetByHash(hashOpaqueToken("raw-token")).
		Return(&repository.EmailActionToken{
			UserId:    "user-uuid-test",
			Action:    repository.EmailActionResetPassword,
			ExpiresAt: time.Now().Add(time.Minute),
		}, nil)
	svc.EmailActionTokenRepository.EXPECT().
		DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionResetPassword).
		Return(nil)
	svc.UserRepository.EXPECT().
		SetPassword("user-uuid-test", gomock.Any()).
		DoAndReturn(func(_, hashed string) error {
			require.NoError(t, checkPassword(hashed, "brand-new-password"))
			return nil
		})
	svc.RefreshTokenRepository.EXPECT().DeleteByUserId("user-uuid-test").Return(nil)

	err := svc.Service.EmailVerificationService.ResetPassword("raw-token", "brand-new-password")

	require.NoError(t, err)
}

func Test_Unit_PasswordReset_Reset_RejectsBadTokens(t *testing.T) {
	cases := map[string]*repository.EmailActionToken{
		"unknown": nil,
		"expired": {UserId: "u", Action: repository.EmailActionResetPassword, ExpiresAt: time.Now().Add(-time.Minute)},
		// A verify or invite link must never be usable to change a password.
		"verify link": {UserId: "u", Action: repository.EmailActionVerify, ExpiresAt: time.Now().Add(time.Hour)},
		"invite link": {UserId: "u", Action: repository.EmailActionSetPassword, ExpiresAt: time.Now().Add(time.Hour)},
	}
	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			svc := NewMockService(t)
			defer svc.Ctrl.Finish()
			setupPasswordResetTestConfig(t, true)

			svc.EmailActionTokenRepository.EXPECT().GetByHash(gomock.Any()).Return(token, nil)

			err := svc.Service.EmailVerificationService.ResetPassword("raw-token", "brand-new-password")

			require.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}

func Test_Unit_PasswordReset_Reset_ANormalSetPasswordDoesNotAcceptAResetLink(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, true)

	svc.EmailActionTokenRepository.EXPECT().GetByHash(gomock.Any()).Return(&repository.EmailActionToken{
		UserId: "u", Action: repository.EmailActionResetPassword, ExpiresAt: time.Now().Add(time.Hour),
	}, nil)

	err := svc.Service.EmailVerificationService.SetPassword("raw-token", "brand-new-password")

	require.ErrorIs(t, err, ErrInvalidToken)
}

func Test_Unit_PasswordReset_Reset_Disabled(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupPasswordResetTestConfig(t, false)

	err := svc.Service.EmailVerificationService.ResetPassword("raw-token", "brand-new-password")

	require.ErrorIs(t, err, ErrPasswordResetDisabled)
}

func Test_Unit_PasswordReset_Reset_PropagatesFailures(t *testing.T) {
	token := &repository.EmailActionToken{UserId: "user-uuid-test", Action: repository.EmailActionResetPassword, ExpiresAt: time.Now().Add(time.Minute)}

	t.Run("setting the password", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		setupPasswordResetTestConfig(t, true)
		svc.EmailActionTokenRepository.EXPECT().GetByHash(gomock.Any()).Return(token, nil)
		svc.EmailActionTokenRepository.EXPECT().DeleteByUserIdAndAction(gomock.Any(), gomock.Any()).Return(nil)
		svc.UserRepository.EXPECT().SetPassword("user-uuid-test", gomock.Any()).Return(errors.New("db unavailable"))

		require.ErrorContains(t, svc.Service.EmailVerificationService.ResetPassword("t", "brand-new-password"), "db unavailable")
	})

	t.Run("ending the sessions", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		setupPasswordResetTestConfig(t, true)
		svc.EmailActionTokenRepository.EXPECT().GetByHash(gomock.Any()).Return(token, nil)
		svc.EmailActionTokenRepository.EXPECT().DeleteByUserIdAndAction(gomock.Any(), gomock.Any()).Return(nil)
		svc.UserRepository.EXPECT().SetPassword("user-uuid-test", gomock.Any()).Return(nil)
		svc.RefreshTokenRepository.EXPECT().DeleteByUserId("user-uuid-test").Return(errors.New("db unavailable"))

		require.ErrorContains(t, svc.Service.EmailVerificationService.ResetPassword("t", "brand-new-password"), "db unavailable")
	})
}
