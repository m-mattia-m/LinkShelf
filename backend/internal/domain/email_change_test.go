package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/mailer"
	"backend/internal/infrastructure/repository"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Email changes must not keep the old address's verified status (OIDC pre-hijack).

func setupEmailChangeTestConfig(t *testing.T, verificationEnabled bool) {
	t.Helper()
	setupEmailVerificationTestConfig(t)
	t.Cleanup(config.Reset)
	config.Set("authentication.emailVerification.enabled", verificationEnabled)
}

func existingVerifiedUser() *model.User {
	return &model.User{
		Id:            "user-uuid-test",
		UserBase:      model.UserBase{Email: "old@test.com", Username: "user", Role: model.RoleUser},
		EmailVerified: true,
	}
}

func Test_Unit_User_Update_EmailChange_IsPendingUntilConfirmed(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailChangeTestConfig(t, true)

	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)
	svc.UserRepository.EXPECT().FindByEmail("new@test.com").Return(nil, nil)
	svc.EmailActionTokenRepository.EXPECT().
		GetLatestByUserIdAndAction("user-uuid-test", repository.EmailActionChangeEmail).
		Return(nil, nil)

	// Older links must be deleted before pending_email changes.
	gomock.InOrder(
		svc.EmailActionTokenRepository.EXPECT().
			DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionChangeEmail).
			Return(nil),
		svc.UserRepository.EXPECT().SetPendingEmail("user-uuid-test", "new@test.com").Return(nil),
		svc.EmailActionTokenRepository.EXPECT().
			Create("user-uuid-test", gomock.Any(), repository.EmailActionChangeEmail, gomock.Any()).
			Return(nil),
	)

	var sentTo []string
	svc.Mailer.EXPECT().Send(gomock.Any()).Times(2).DoAndReturn(func(msg mailer.Message) error {
		sentTo = append(sentTo, msg.To)
		if msg.To == "new@test.com" {
			require.Contains(t, msg.TextBody, "/auth/confirm-email?token=")
		} else {
			require.Contains(t, msg.TextBody, "new@test.com")
			require.NotContains(t, msg.TextBody, "token=")
		}
		return nil
	})

	// The profile is saved with the CURRENT email - the new one only waits.
	svc.UserRepository.EXPECT().
		Update(gomock.Any()).
		DoAndReturn(func(u *model.User) error {
			require.Equal(t, "old@test.com", u.Email)
			require.Equal(t, "New", u.FirstName)
			return nil
		})
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(&model.User{
		Id:            "user-uuid-test",
		UserBase:      model.UserBase{Email: "old@test.com"},
		EmailVerified: true,
		PendingEmail:  "new@test.com",
	}, nil)

	updated, err := svc.Service.UserService.Update("user-uuid-test", &model.User{
		UserBase: model.UserBase{Email: "  New@Test.com ", Username: "user", FirstName: "New"},
	}, false)

	require.NoError(t, err)
	require.Equal(t, "old@test.com", updated.Email)
	require.Equal(t, "new@test.com", updated.PendingEmail)
	require.False(t, updated.EmailDeliveryFailed)
	require.ElementsMatch(t, []string{"new@test.com", "old@test.com"}, sentTo)
}

// Q16: an admin changing someone else's email goes through the same flow.
func Test_Unit_User_Update_EmailChange_ByAdmin_IsPendingToo(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailChangeTestConfig(t, true)

	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)
	svc.UserRepository.EXPECT().FindByEmail("new@test.com").Return(nil, nil)
	svc.EmailActionTokenRepository.EXPECT().GetLatestByUserIdAndAction(gomock.Any(), gomock.Any()).Return(nil, nil)
	svc.EmailActionTokenRepository.EXPECT().DeleteByUserIdAndAction(gomock.Any(), gomock.Any()).Return(nil)
	svc.UserRepository.EXPECT().SetPendingEmail("user-uuid-test", "new@test.com").Return(nil)
	svc.EmailActionTokenRepository.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	svc.Mailer.EXPECT().Send(gomock.Any()).Times(2).Return(nil)
	svc.UserRepository.EXPECT().
		Update(gomock.Any()).
		DoAndReturn(func(u *model.User) error {
			require.Equal(t, "old@test.com", u.Email)
			return nil
		})
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)

	_, err := svc.Service.UserService.Update("user-uuid-test", &model.User{
		UserBase: model.UserBase{Email: "new@test.com", Username: "user"},
	}, true)

	require.NoError(t, err)
}

func Test_Unit_User_Update_EmailChange_ReportsAFailedDelivery(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailChangeTestConfig(t, true)

	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)
	svc.UserRepository.EXPECT().FindByEmail("new@test.com").Return(nil, nil)
	svc.EmailActionTokenRepository.EXPECT().GetLatestByUserIdAndAction(gomock.Any(), gomock.Any()).Return(nil, nil)
	svc.EmailActionTokenRepository.EXPECT().DeleteByUserIdAndAction(gomock.Any(), gomock.Any()).Return(nil)
	svc.UserRepository.EXPECT().SetPendingEmail("user-uuid-test", "new@test.com").Return(nil)
	svc.EmailActionTokenRepository.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	svc.Mailer.EXPECT().Send(gomock.Any()).Times(2).Return(errors.New("smtp down"))
	svc.UserRepository.EXPECT().Update(gomock.Any()).Return(nil)
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)

	updated, err := svc.Service.UserService.Update("user-uuid-test", &model.User{
		UserBase: model.UserBase{Email: "new@test.com", Username: "user"},
	}, false)

	require.NoError(t, err)
	require.True(t, updated.EmailDeliveryFailed)
}

func Test_Unit_User_Update_EmailChange_RefusesAnEmailOfAnotherAccount(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailChangeTestConfig(t, true)

	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)
	svc.UserRepository.EXPECT().
		FindByEmail("taken@test.com").
		Return(&repository.AuthRecord{Id: "someone-else"}, nil)
	// Nothing is written: the mock fails on Update/SetPendingEmail.

	updated, err := svc.Service.UserService.Update("user-uuid-test", &model.User{
		UserBase: model.UserBase{Email: "Taken@test.com", Username: "user"},
	}, false)

	require.ErrorIs(t, err, ErrConflict)
	require.Nil(t, updated)
}

func Test_Unit_User_Update_EmailChange_RateLimited_WritesNothing(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailChangeTestConfig(t, true)

	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)
	svc.UserRepository.EXPECT().FindByEmail("new@test.com").Return(nil, nil)
	svc.EmailActionTokenRepository.EXPECT().
		GetLatestByUserIdAndAction("user-uuid-test", repository.EmailActionChangeEmail).
		Return(&repository.EmailActionToken{CreatedAt: time.Now()}, nil)

	updated, err := svc.Service.UserService.Update("user-uuid-test", &model.User{
		UserBase: model.UserBase{Email: "new@test.com", Username: "user"},
	}, false)

	require.ErrorIs(t, err, ErrTooManyRequests)
	require.Nil(t, updated)
}

// Without email verification the change is applied, but unverified.
func Test_Unit_User_Update_EmailChange_WithoutVerification_AppliesItUnverified(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailChangeTestConfig(t, false)

	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)
	svc.UserRepository.EXPECT().FindByEmail("new@test.com").Return(nil, nil)
	gomock.InOrder(
		svc.UserRepository.EXPECT().
			Update(gomock.Any()).
			DoAndReturn(func(u *model.User) error {
				require.Equal(t, "old@test.com", u.Email)
				return nil
			}),
		svc.UserRepository.EXPECT().ChangeEmail("user-uuid-test", "new@test.com", false).Return(nil),
	)
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(&model.User{
		Id: "user-uuid-test", UserBase: model.UserBase{Email: "new@test.com"}, EmailVerified: false,
	}, nil)

	updated, err := svc.Service.UserService.Update("user-uuid-test", &model.User{
		UserBase: model.UserBase{Email: "new@test.com", Username: "user"},
	}, false)

	require.NoError(t, err)
	require.Equal(t, "new@test.com", updated.Email)
	require.False(t, updated.EmailVerified)
}

// Only the case differs: that is the same address, not a change.
func Test_Unit_User_Update_SameEmailInOtherCase_IsNoChange(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailChangeTestConfig(t, true)

	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)
	svc.UserRepository.EXPECT().Update(gomock.Any()).Return(nil)
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(existingVerifiedUser(), nil)

	_, err := svc.Service.UserService.Update("user-uuid-test", &model.User{
		UserBase: model.UserBase{Email: "OLD@test.com", Username: "user"},
	}, false)

	require.NoError(t, err)
}

func Test_Unit_EmailChange_Confirm_AppliesThePendingEmailAsVerified(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.EmailActionTokenRepository.EXPECT().GetByHash(gomock.Any()).Return(&repository.EmailActionToken{
		UserId: "user-uuid-test", Action: repository.EmailActionChangeEmail, ExpiresAt: time.Now().Add(time.Hour),
	}, nil)
	svc.EmailActionTokenRepository.EXPECT().DeleteByUserIdAndAction("user-uuid-test", repository.EmailActionChangeEmail).Return(nil)
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(&model.User{
		Id: "user-uuid-test", UserBase: model.UserBase{Email: "old@test.com"}, PendingEmail: "new@test.com",
	}, nil)
	svc.UserRepository.EXPECT().FindByEmail("new@test.com").Return(nil, nil)
	svc.UserRepository.EXPECT().ChangeEmail("user-uuid-test", "new@test.com", true).Return(nil)

	require.NoError(t, svc.Service.EmailVerificationService.ConfirmEmailChange("raw-token"))
}

func Test_Unit_EmailChange_Confirm_FailsWhenTheAddressWasTakenMeanwhile(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.EmailActionTokenRepository.EXPECT().GetByHash(gomock.Any()).Return(&repository.EmailActionToken{
		UserId: "user-uuid-test", Action: repository.EmailActionChangeEmail, ExpiresAt: time.Now().Add(time.Hour),
	}, nil)
	svc.EmailActionTokenRepository.EXPECT().DeleteByUserIdAndAction(gomock.Any(), gomock.Any()).Return(nil)
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(&model.User{
		Id: "user-uuid-test", PendingEmail: "new@test.com",
	}, nil)
	svc.UserRepository.EXPECT().FindByEmail("new@test.com").Return(&repository.AuthRecord{Id: "someone-else"}, nil)
	svc.UserRepository.EXPECT().SetPendingEmail("user-uuid-test", "").Return(nil)

	err := svc.Service.EmailVerificationService.ConfirmEmailChange("raw-token")

	require.ErrorIs(t, err, ErrConflict)
}

func Test_Unit_EmailChange_Confirm_RejectsATokenWithoutPendingEmail(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.EmailActionTokenRepository.EXPECT().GetByHash(gomock.Any()).Return(&repository.EmailActionToken{
		UserId: "user-uuid-test", Action: repository.EmailActionChangeEmail, ExpiresAt: time.Now().Add(time.Hour),
	}, nil)
	svc.EmailActionTokenRepository.EXPECT().DeleteByUserIdAndAction(gomock.Any(), gomock.Any()).Return(nil)
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(&model.User{Id: "user-uuid-test"}, nil)

	err := svc.Service.EmailVerificationService.ConfirmEmailChange("raw-token")

	require.ErrorIs(t, err, ErrInvalidToken)
}

// Tokens are bound to their action and expire.
func Test_Unit_EmailChange_Confirm_RejectsOtherAndExpiredTokens(t *testing.T) {
	for name, token := range map[string]*repository.EmailActionToken{
		"verify token":  {UserId: "user-uuid-test", Action: repository.EmailActionVerify, ExpiresAt: time.Now().Add(time.Hour)},
		"expired token": {UserId: "user-uuid-test", Action: repository.EmailActionChangeEmail, ExpiresAt: time.Now().Add(-time.Minute)},
		"unknown token": nil,
	} {
		t.Run(name, func(t *testing.T) {
			svc := NewMockService(t)
			defer svc.Ctrl.Finish()

			svc.EmailActionTokenRepository.EXPECT().GetByHash(gomock.Any()).Return(token, nil)

			err := svc.Service.EmailVerificationService.ConfirmEmailChange("raw-token")

			require.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}

func Test_Unit_EmailChange_NoticeEscapesTheNewAddressInHtml(t *testing.T) {
	setupEmailVerificationTestConfig(t)

	msg := emailChangeRequestedMessage("old@test.com", `"><script>alert(1)</script>@evil.tld`)

	require.NotContains(t, msg.HTMLBody, "<script>")
	require.Contains(t, msg.HTMLBody, "&lt;script&gt;")
}

func Test_Unit_User_Creation_StoresTheEmailLowercased(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	allowSelfRegistration(t)

	svc.UserRepository.EXPECT().UsernameTaken("new-user", "").Return(false, nil)
	svc.UserRepository.EXPECT().
		Create(gomock.Any(), gomock.Any(), model.RoleUser, gomock.Any()).
		DoAndReturn(func(u model.UserBase, _, _ string, _ *int) (string, error) {
			require.Equal(t, "new.user@test.com", u.Email)
			return "user-uuid-test", nil
		})
	svc.UserRepository.EXPECT().Get("user-uuid-test").Return(&model.User{Id: "user-uuid-test"}, nil)

	_, err := svc.Service.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Email: " New.User@Test.COM ", Username: "new-user"},
		Password: "a-password",
	}, false)

	require.NoError(t, err)
}

func Test_Unit_User_Creation_SelfRegistrationRefusedWhenLocalAuthIsDisabled(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	allowSelfRegistration(t)
	config.Set("authentication.localAuthEnabled", false)

	user, err := svc.Service.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{Email: "new@test.com", Username: "new-user"},
		Password: "a-password",
	}, false)

	require.ErrorIs(t, err, ErrRegistrationDisabled)
	require.Nil(t, user)
}

func Test_Unit_PasswordReset_RefusedWhenLocalAuthIsDisabled(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	setupEmailVerificationTestConfig(t)
	t.Cleanup(config.Reset)
	config.Set("authentication.passwordReset.enabled", true)
	config.Set("authentication.localAuthEnabled", false)

	require.ErrorIs(t, svc.Service.EmailVerificationService.RequestPasswordReset("user@test.com"), ErrPasswordResetDisabled)
	require.ErrorIs(t, svc.Service.EmailVerificationService.ResetPassword("raw-token", "new-password"), ErrPasswordResetDisabled)
}
