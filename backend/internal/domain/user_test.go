package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// allowSelfRegistration is needed by every test that creates a user as a
// non-admin caller, since Create() checks
// authentication.registrationEnabled for that path.
func allowSelfRegistration(t *testing.T) {
	t.Helper()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("authentication.registrationEnabled", true)
}

func Test_Unit_User_List_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		List().
		Return([]model.User{
			{Id: "user-uuid-test", UserBase: model.UserBase{Email: "test@test.com"}},
		}, nil)

	users, err := svc.Service.UserService.List()

	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, "user-uuid-test", users[0].Id)
}

func Test_Unit_User_List_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		List().
		Return(nil, errors.New("an error occurred"))

	users, err := svc.Service.UserService.List()

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, users)
}

func Test_Unit_User_Creation_Success_SelfRegistration_DefaultsToUserRole(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	allowSelfRegistration(t)

	svc.UserRepository.
		EXPECT().
		UsernameTaken("test-user", "").
		Return(false, nil)

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any(), gomock.Any(), model.RoleUser).
		Return("user-uuid-test", nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{
			Id: "user-uuid-test",
			UserBase: model.UserBase{
				Username:  "test-user",
				FirstName: "firstname-test",
				LastName:  "lastname-test",
				Email:     "test@test.com",
				Role:      model.RoleUser,
			},
		}, nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Username: "test-user",
			Email:     "test@test.com",
			FirstName: "firstname-test",
			LastName:  "lastname-test",
			// A non-admin caller attempting to self-elevate must be ignored.
			Role: model.RoleAdmin,
		},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, false)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, model.RoleUser, user.Role)
}

func Test_Unit_User_Creation_Failure_RegistrationDisabled(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("authentication.registrationEnabled", false)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Email: "test@test.com", FirstName: "First", LastName: "Last"},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, false)

	require.ErrorIs(t, err, ErrRegistrationDisabled)
	require.Nil(t, user)
}

func Test_Unit_User_Creation_Success_AdminBypassesRegistrationDisabled(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("authentication.registrationEnabled", false)

	svc.UserRepository.
		EXPECT().
		UsernameTaken("test-user", "").
		Return(false, nil)

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any(), gomock.Any(), model.RoleUser).
		Return("user-uuid-test", nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{
			Username: "test-user", Role: model.RoleUser}}, nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Username: "test-user", Email: "test@test.com", FirstName: "First", LastName: "Last"},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, true)

	require.NoError(t, err)
	require.NotNil(t, user)
}

func Test_Unit_User_Creation_Failure_SelfRegistrationRequiresPassword(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	allowSelfRegistration(t)

	svc.UserRepository.
		EXPECT().
		UsernameTaken("test-user", "").
		Return(false, nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Username: "test-user", Email: "test@test.com", FirstName: "First", LastName: "Last"},
		Password: "",
	}
	user, err := svc.Service.UserService.Create(&userRequest, false)

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, user)
}

func Test_Unit_User_Creation_Failure_AdminRequiresPasswordWhenVerificationDisabled(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("authentication.emailVerification.enabled", false)

	svc.UserRepository.
		EXPECT().
		UsernameTaken("test-user", "").
		Return(false, nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Username: "test-user", Email: "test@test.com", FirstName: "First", LastName: "Last"},
		Password: "",
	}
	user, err := svc.Service.UserService.Create(&userRequest, true)

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, user)
}

func Test_Unit_User_Creation_Success_AdminInvitesPasswordlessAccount(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("authentication.emailVerification.enabled", true)
	config.Set("authentication.emailVerification.tokenExpiryHours", 24)

	svc.UserRepository.
		EXPECT().
		UsernameTaken("test-user", "").
		Return(false, nil)

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any(), "", model.RoleUser).
		Return("invited-uuid-test", nil)

	svc.UserRepository.
		EXPECT().
		Get("invited-uuid-test").
		Return(&model.User{
			Id:          "invited-uuid-test",
			UserBase:    model.UserBase{Email: "invited@test.com", Role: model.RoleUser},
			HasPassword: false,
		}, nil)

	svc.EmailActionTokenRepository.
		EXPECT().
		GetLatestByUserIdAndAction("invited-uuid-test", repository.EmailActionSetPassword).
		Return(nil, nil)
	svc.EmailActionTokenRepository.
		EXPECT().
		DeleteByUserIdAndAction("invited-uuid-test", repository.EmailActionSetPassword).
		Return(nil)
	svc.EmailActionTokenRepository.
		EXPECT().
		Create("invited-uuid-test", gomock.Any(), repository.EmailActionSetPassword, gomock.Any()).
		Return(nil)
	svc.Mailer.EXPECT().Send(gomock.Any()).Return(nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{
			Username: "test-user", Email: "invited@test.com", FirstName: "First", LastName: "Last"},
		Password: "",
	}
	user, err := svc.Service.UserService.Create(&userRequest, true)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.False(t, user.HasPassword)
}

func Test_Unit_User_Creation_Success_AdminSetsRole(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	config.Reset()

	svc.UserRepository.
		EXPECT().
		UsernameTaken("test-user", "").
		Return(false, nil)

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any(), gomock.Any(), model.RoleAdmin).
		Return("user-uuid-test", nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{
			Id: "user-uuid-test",
			UserBase: model.UserBase{
				Username: "test-user", Role: model.RoleAdmin},
		}, nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Username: "test-user", Role: model.RoleAdmin},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, true)

	require.NoError(t, err)
	require.Equal(t, model.RoleAdmin, user.Role)
}

func Test_Unit_User_Creation_Failure_AdminSetsInvalidRole(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Role: "superuser"},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, true)

	require.ErrorIs(t, err, ErrInvalidRole)
	require.Nil(t, user)
}

func Test_Unit_User_Creation_Failure_Creation(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	allowSelfRegistration(t)

	svc.UserRepository.
		EXPECT().
		UsernameTaken("test-user", "").
		Return(false, nil)

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errors.New("an error occurred"))

	userRequest := model.UserCreate{
		UserBase: model.UserBase{
			Username:  "test-user",
			Email:     "test@test.com",
			FirstName: "firstname-test",
			LastName:  "lastname-test",
		},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, false)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, user)
}

func Test_Unit_User_Creation_Failure_Get(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	allowSelfRegistration(t)

	svc.UserRepository.
		EXPECT().
		UsernameTaken("test-user", "").
		Return(false, nil)

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("user-uuid-test", nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	userRequest := model.UserCreate{
		UserBase: model.UserBase{
			Username:  "test-user",
			Email:     "test@test.com",
			FirstName: "firstname-test",
			LastName:  "lastname-test",
		},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, false)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, user)
}

func Test_Unit_User_Get_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userRequest := model.User{
		Id: "user-uuid-test",
		UserBase: model.UserBase{
			FirstName: "firstname-test",
			LastName:  "lastname-test",
			Email:     "test@test.com",
		},
	}
	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&userRequest, nil)

	user, err := svc.Service.UserService.Get("user-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "user-uuid-test", user.Id)
	require.Equal(t, "firstname-test", user.FirstName)
	require.Equal(t, "lastname-test", user.LastName)
	require.Equal(t, "test@test.com", user.Email)
}

func Test_Unit_User_Get_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	user, err := svc.Service.UserService.Get("user-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, user)
}

func Test_Unit_User_Update_Success_SelfUpdate_RoleUnchanged(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userRequest := model.User{
		UserBase: model.UserBase{
			FirstName: "firstname-updated-test",
			LastName:  "lastname-updated-test",
			Email:     "test@test.com",
			// A non-admin caller attempting to self-elevate must be ignored.
			Role: model.RoleAdmin,
		},
	}

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{Role: model.RoleUser}}, nil)

	svc.UserRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{
			Id: "user-uuid-test",
			UserBase: model.UserBase{
				FirstName: "firstname-updated-test",
				LastName:  "lastname-updated-test",
				Email:     "test@test.com",
				Role:      model.RoleUser,
			},
		}, nil)

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &userRequest, false)

	require.NoError(t, err)
	require.NotNil(t, updatedUser)
	require.Equal(t, "user-uuid-test", updatedUser.Id)
	require.Equal(t, model.RoleUser, updatedUser.Role)
}

func Test_Unit_User_Update_Success_AdminChangesRole(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userRequest := model.User{UserBase: model.UserBase{Role: model.RoleAdmin}}

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{Role: model.RoleUser}}, nil)

	svc.UserRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{Role: model.RoleAdmin}}, nil)

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &userRequest, true)

	require.NoError(t, err)
	require.Equal(t, model.RoleAdmin, updatedUser.Role)
}

func Test_Unit_User_Update_Failure_AdminSetsInvalidRole(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{Role: model.RoleUser}}, nil)

	userRequest := model.User{UserBase: model.UserBase{Role: "superuser"}}
	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &userRequest, true)

	require.ErrorIs(t, err, ErrInvalidRole)
	require.Nil(t, updatedUser)
}

func Test_Unit_User_Update_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, nil)

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &model.User{}, false)

	require.NoError(t, err)
	require.Nil(t, updatedUser)
}

func Test_Unit_User_Update_Failure_Get_Existing(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &model.User{}, false)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, updatedUser)
}

func Test_Unit_User_Update_Failure_Update(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{Role: model.RoleUser}}, nil)

	svc.UserRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(errors.New("an error occurred"))

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &model.User{}, false)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, updatedUser)
}

func Test_Unit_User_Update_Failure_Get_Final(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{Role: model.RoleUser}}, nil)

	svc.UserRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &model.User{}, false)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, updatedUser)
}

func Test_Unit_User_PatchPassword_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	passwordRequest := model.UserRequestBodyOnlyPassword{
		OldPassword: "secret",
		NewPassword: "new-secret",
	}

	hashedPassword, err := hashPassword("secret")
	require.NoError(t, err)

	svc.UserRepository.
		EXPECT().
		GetPassword("user-uuid-test").
		Return(hashedPassword, nil)

	svc.UserRepository.
		EXPECT().
		PatchPassword(gomock.Any(), gomock.Any()).
		Return(nil)

	err = svc.Service.UserService.PatchPassword("user-uuid-test", &passwordRequest)

	require.NoError(t, err)
}

func Test_Unit_User_PatchPassword_Failure_WrongOldPassword(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	passwordRequest := model.UserRequestBodyOnlyPassword{
		OldPassword: "wrong-secret",
		NewPassword: "new-secret",
	}

	hashedPassword, err := hashPassword("secret")
	require.NoError(t, err)

	svc.UserRepository.
		EXPECT().
		GetPassword("user-uuid-test").
		Return(hashedPassword, nil)

	err = svc.Service.UserService.PatchPassword("user-uuid-test", &passwordRequest)

	require.ErrorContains(t, err, "crypto/bcrypt: hashedPassword is not the hash of the given password")
}

func Test_Unit_User_PatchPassword_Failure_UnequalOldPassword(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	hashedPassword, err := hashPassword("secret")
	require.NoError(t, err)

	passwordRequest := model.UserRequestBodyOnlyPassword{
		OldPassword: "",
		NewPassword: "new-secret",
	}

	svc.UserRepository.
		EXPECT().
		GetPassword("user-uuid-test").
		Return(hashedPassword, nil)

	err = svc.Service.UserService.PatchPassword("user-uuid-test", &passwordRequest)

	require.ErrorContains(t, err, "crypto/bcrypt: hashedPassword is not the hash of the given password")
}

func Test_Unit_User_PatchPassword_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	passwordRequest := model.UserRequestBodyOnlyPassword{
		OldPassword: "secret",
		NewPassword: "new-secret",
	}

	hashedPassword, err := hashPassword("secret")
	require.NoError(t, err)

	svc.UserRepository.
		EXPECT().
		GetPassword("user-uuid-test").
		Return(hashedPassword, nil)

	svc.UserRepository.
		EXPECT().
		PatchPassword(gomock.Any(), gomock.Any()).
		Return(errors.New("an error occurred"))

	err = svc.Service.UserService.PatchPassword("user-uuid-test", &passwordRequest)

	require.ErrorContains(t, err, "an error occurred")
}

func Test_Unit_User_Delete_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userToDelete := &model.User{
		Id: "user-uuid-test",
	}

	svc.UserRepository.
		EXPECT().
		Delete(userToDelete).
		Return(nil)

	err := svc.Service.UserService.Delete(userToDelete)

	require.NoError(t, err)
}

func Test_Unit_User_Delete_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userToDelete := &model.User{
		Id: "user-uuid-test",
	}

	svc.UserRepository.
		EXPECT().
		Delete(userToDelete).
		Return(errors.New("an error occurred"))

	err := svc.Service.UserService.Delete(userToDelete)

	require.ErrorContains(t, err, "an error occurred")
}

func existingUserWithUsername(username string) *model.User {
	return &model.User{
		Id:       "user-1",
		UserBase: model.UserBase{Email: "a@test.com", Username: username, FirstName: "First", LastName: "Last", Role: model.RoleUser},
	}
}

func Test_Unit_User_Creation_Failure_InvalidUsername(t *testing.T) {
	for name, username := range map[string]string{
		"empty":    "",
		"reserved": "admin",
		"route":    "docs",
		"format":   "Not Valid",
	} {
		t.Run(name, func(t *testing.T) {
			svc := NewMockService(t)
			defer svc.Ctrl.Finish()
			allowSelfRegistration(t)

			// No repository expectations: an invalid username never reaches the database.
			userRequest := model.UserCreate{
				UserBase: model.UserBase{Username: username, Email: "a@test.com", FirstName: "First", LastName: "Last"},
				Password: "secret",
			}
			user, err := svc.Service.UserService.Create(&userRequest, false)

			require.ErrorIs(t, err, ErrInvalidInput)
			require.Nil(t, user)
		})
	}
}

func Test_Unit_User_Creation_Failure_UsernameTaken(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	allowSelfRegistration(t)

	svc.UserRepository.EXPECT().UsernameTaken("test-user", "").Return(true, nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Username: "test-user", Email: "a@test.com", FirstName: "First", LastName: "Last"},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, false)

	require.ErrorIs(t, err, ErrConflict)
	require.Nil(t, user)
}

func Test_Unit_User_Creation_Failure_UsernameLookupFails(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	allowSelfRegistration(t)

	svc.UserRepository.EXPECT().UsernameTaken("test-user", "").Return(false, errors.New("db unavailable"))

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Username: "test-user", Email: "a@test.com", FirstName: "First", LastName: "Last"},
		Password: "secret",
	}
	user, err := svc.Service.UserService.Create(&userRequest, false)

	require.ErrorContains(t, err, "db unavailable")
	require.Nil(t, user)
}

func Test_Unit_User_Update_Success_RenameChecksTheNewUsername(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.EXPECT().Get("user-1").Return(existingUserWithUsername("old-name"), nil).Times(2)
	svc.UserRepository.EXPECT().UsernameTaken("new-name", "user-1").Return(false, nil)
	svc.UserRepository.EXPECT().Update(gomock.Any()).DoAndReturn(func(u *model.User) error {
		require.Equal(t, "new-name", u.Username)
		return nil
	})

	request := model.User{UserBase: model.UserBase{Email: "a@test.com", Username: "new-name", FirstName: "First", LastName: "Last"}}
	updated, err := svc.Service.UserService.Update("user-1", &request, false)

	require.NoError(t, err)
	require.NotNil(t, updated)
}

func Test_Unit_User_Update_Failure_RenameToTakenUsername(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.EXPECT().Get("user-1").Return(existingUserWithUsername("old-name"), nil)
	svc.UserRepository.EXPECT().UsernameTaken("taken-name", "user-1").Return(true, nil)

	request := model.User{UserBase: model.UserBase{Email: "a@test.com", Username: "taken-name", FirstName: "First", LastName: "Last"}}
	updated, err := svc.Service.UserService.Update("user-1", &request, false)

	require.ErrorIs(t, err, ErrConflict)
	require.Nil(t, updated)
}

func Test_Unit_User_Update_Failure_RenameToInvalidOrReservedUsername(t *testing.T) {
	for _, username := range []string{"admin", "docs", "ab", "Bad Name", ""} {
		t.Run(username, func(t *testing.T) {
			svc := NewMockService(t)
			defer svc.Ctrl.Finish()

			svc.UserRepository.EXPECT().Get("user-1").Return(existingUserWithUsername("old-name"), nil)

			request := model.User{UserBase: model.UserBase{Email: "a@test.com", Username: username, FirstName: "First", LastName: "Last"}}
			updated, err := svc.Service.UserService.Update("user-1", &request, false)

			require.ErrorIs(t, err, ErrInvalidInput)
			require.Nil(t, updated)
		})
	}
}

func Test_Unit_User_Update_Success_UnchangedUsernameIsNotChecked(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	// "admin" is reserved, but it is what this account already has (the
	// bootstrap admin), so saving other profile fields must still work and
	// must not ask the repository about it.
	svc.UserRepository.EXPECT().Get("user-1").Return(existingUserWithUsername("admin"), nil).Times(2)
	svc.UserRepository.EXPECT().Update(gomock.Any()).Return(nil)

	request := model.User{UserBase: model.UserBase{Email: "a@test.com", Username: "admin", FirstName: "New", LastName: "Last"}}
	updated, err := svc.Service.UserService.Update("user-1", &request, false)

	require.NoError(t, err)
	require.NotNil(t, updated)
}

func Test_Unit_User_Update_Failure_UsernameLookupFails(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.EXPECT().Get("user-1").Return(existingUserWithUsername("old-name"), nil)
	svc.UserRepository.EXPECT().UsernameTaken("new-name", "user-1").Return(false, errors.New("db unavailable"))

	request := model.User{UserBase: model.UserBase{Email: "a@test.com", Username: "new-name", FirstName: "First", LastName: "Last"}}
	updated, err := svc.Service.UserService.Update("user-1", &request, false)

	require.ErrorContains(t, err, "db unavailable")
	require.Nil(t, updated)
}
