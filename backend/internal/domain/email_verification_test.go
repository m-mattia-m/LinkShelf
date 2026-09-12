package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/mailer"
	"backend/internal/infrastructure/repository"
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
