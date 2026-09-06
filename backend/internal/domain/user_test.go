package domain

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
				FirstName: "firstname-test",
				LastName:  "lastname-test",
				Email:     "test@test.com",
				Role:      model.RoleUser,
			},
		}, nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{
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

func Test_Unit_User_Creation_Success_AdminSetsRole(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any(), gomock.Any(), model.RoleAdmin).
		Return("user-uuid-test", nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{
			Id:       "user-uuid-test",
			UserBase: model.UserBase{Role: model.RoleAdmin},
		}, nil)

	userRequest := model.UserCreate{
		UserBase: model.UserBase{Role: model.RoleAdmin},
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

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errors.New("an error occurred"))

	userRequest := model.UserCreate{
		UserBase: model.UserBase{
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
