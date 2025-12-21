package controller

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_API_CreateUser_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateUser(svc.Service)

	input := &model.UserRequestBody{
		Body: model.UserBase{
			Email:     "test@test.com",
			FirstName: "firstname",
			LastName:  "lastname",
		},
	}

	svc.UserService.
		EXPECT().
		Create(gomock.Any()).
		Return(&model.User{
			Id: "user-uuid-test",
			UserBase: model.UserBase{
				Email:     "test@test.com",
				FirstName: "firstname",
				LastName:  "lastname",
			},
		}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "user-uuid-test", resp.Body.Id)
	require.Equal(t, "test@test.com", resp.Body.Email)
}

func Test_API_CreateUser_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateUser(svc.Service)

	svc.UserService.
		EXPECT().
		Create(gomock.Any()).
		Return(nil, errors.New("failed to create user"))

	resp, err := handler(context.Background(), &model.UserRequestBody{})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to create user")
}

func Test_API_GetUserById_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetUserById(svc.Service)

	svc.UserService.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{
			Id: "user-uuid-test",
			UserBase: model.UserBase{
				Email: "test@test.com",
			},
		}, nil)

	resp, err := handler(context.Background(), &model.UserRequestFilter{
		UserId: "user-uuid-test",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "user-uuid-test", resp.Body.Id)
}

func Test_API_GetUserById_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetUserById(svc.Service)

	svc.UserService.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("failed to get user"))

	resp, err := handler(context.Background(), &model.UserRequestFilter{
		UserId: "user-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to get user")
}

func Test_API_UpdateUser_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateUser(svc.Service)

	input := &model.UserFilterFilterAndBody{
		UserRequestFilter: model.UserRequestFilter{
			UserId: "user-uuid-test",
		},
		Body: model.UserBase{
			FirstName: "updated-firstname",
		},
	}

	svc.UserService.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(&model.User{
			Id: "user-uuid-test",
			UserBase: model.UserBase{
				FirstName: "updated-firstname",
			},
		}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "updated-firstname", resp.Body.FirstName)
}

func Test_API_UpdateUser_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateUser(svc.Service)

	svc.UserService.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("failed to update user"))

	resp, err := handler(context.Background(), &model.UserFilterFilterAndBody{})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to update user")
}

func Test_API_PatchUserPassword_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := PatchUserPassword(svc.Service)

	svc.UserService.
		EXPECT().
		PatchPassword("user-uuid-test", gomock.Any()).
		Return(nil)

	resp, err := handler(context.Background(), &model.UserPatchPasswordFilterAndBody{
		UserRequestFilter: model.UserRequestFilter{
			UserId: "user-uuid-test",
		},
		Body: model.UserRequestBodyOnlyPassword{
			OldPassword: "secret",
			NewPassword: "new-password",
		},
	})

	require.NoError(t, err)
	require.Nil(t, resp)
}

func Test_API_PatchUserPassword_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := PatchUserPassword(svc.Service)

	svc.UserService.
		EXPECT().
		PatchPassword("user-uuid-test", gomock.Any()).
		Return(errors.New("failed to patch user password"))

	resp, err := handler(context.Background(), &model.UserPatchPasswordFilterAndBody{
		UserRequestFilter: model.UserRequestFilter{
			UserId: "user-uuid-test",
		},
		Body: model.UserRequestBodyOnlyPassword{
			OldPassword: "secret",
			NewPassword: "new-password",
		},
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to patch user password")
}

func Test_API_DeleteUser_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteUser(svc.Service)

	gomock.InOrder(
		svc.UserService.
			EXPECT().
			Get("user-uuid-test").
			Return(&model.User{Id: "user-uuid-test"}, nil),

		svc.UserService.
			EXPECT().
			Delete(gomock.Any()).
			Return(nil),
	)

	resp, err := handler(context.Background(), &model.UserRequestFilter{
		UserId: "user-uuid-test",
	})

	require.NoError(t, err)
	require.Nil(t, resp)
}

func Test_API_DeleteUser_Failure_Get(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteUser(svc.Service)

	svc.UserService.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("failed to get user"))

	resp, err := handler(context.Background(), &model.UserRequestFilter{
		UserId: "user-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to get user")
}

func Test_API_DeleteUser_Failure_Delete(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteUser(svc.Service)

	svc.UserService.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

	svc.UserService.
		EXPECT().
		Delete(gomock.Any()).
		Return(errors.New("failed to delete user"))

	resp, err := handler(context.Background(), &model.UserRequestFilter{
		UserId: "user-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to delete user")
}
