package domain

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Unit_User_Creation_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any()).
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
			},
		}, nil)

	userRequest := model.User{
		UserBase: model.UserBase{
			Email:     "test@test.com",
			FirstName: "firstname-test",
			LastName:  "lastname-test",
			Password:  "secret",
		},
	}
	user, err := svc.Service.UserService.Create(&userRequest)

	require.NoError(t, err)
	require.NotNil(t, user)

	require.NotEmpty(t, user.Id)
	require.Equal(t, userRequest.FirstName, user.FirstName)
	require.Equal(t, userRequest.LastName, user.LastName)
	require.Equal(t, userRequest.Email, user.Email)
	require.NotEqual(t, userRequest.Password, user.Password)

}

func Test_Unit_User_Creation_Failure_Creation(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any()).
		Return("", errors.New("an error occurred"))

	userRequest := model.User{
		UserBase: model.UserBase{
			Email:     "test@test.com",
			FirstName: "firstname-test",
			LastName:  "lastname-test",
			Password:  "secret",
		},
	}
	user, err := svc.Service.UserService.Create(&userRequest)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, user)
}

func Test_Unit_User_Creation_Failure_Get(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Create(gomock.Any()).
		Return("user-uuid-test", nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	userRequest := model.User{
		UserBase: model.UserBase{
			Email:     "test@test.com",
			FirstName: "firstname-test",
			LastName:  "lastname-test",
			Password:  "secret",
		},
	}
	user, err := svc.Service.UserService.Create(&userRequest)

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

func Test_Unit_User_Update_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userRequest := model.User{
		UserBase: model.UserBase{
			FirstName: "firstname-updated-test",
			LastName:  "lastname-updated-test",
			Email:     "test@test.com",
		},
	}

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
			},
		}, nil)

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &userRequest)

	require.NoError(t, err)
	require.NotNil(t, updatedUser)
	require.Equal(t, "user-uuid-test", updatedUser.Id)
	require.Equal(t, "firstname-updated-test", updatedUser.FirstName)
	require.Equal(t, "lastname-updated-test", updatedUser.LastName)
	require.Equal(t, "test@test.com", updatedUser.Email)

}

func Test_Unit_User_Update_Failure_Update(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userRequest := model.User{
		UserBase: model.UserBase{
			FirstName: "firstname-updated-test",
			LastName:  "lastname-updated-test",
			Email:     "test@test.com",
		},
	}

	svc.UserRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(errors.New("an error occurred"))

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &userRequest)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, updatedUser)
}

func Test_Unit_User_Update_Failure_Get(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	userRequest := model.User{
		UserBase: model.UserBase{
			FirstName: "firstname-updated-test",
			LastName:  "lastname-updated-test",
			Email:     "test@test.com",
		},
	}

	svc.UserRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(nil)

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	updatedUser, err := svc.Service.UserService.Update("user-uuid-test", &userRequest)

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
		PatchPassword(gomock.Any()).
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
		PatchPassword(gomock.Any()).
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
