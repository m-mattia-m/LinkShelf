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

func makeTestBearerToken(t *testing.T, subject, role string) string {
	t.Helper()
	config.Reset()
	config.Set("authentication.jwtSecret", "test-secret")

	claims := domain.AccessTokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)
	return signed
}

func Test_API_GetCurrentUser_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetCurrentUser(svc.Service)

	svc.UserService.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{
			Id: "user-uuid-test",
			UserBase: model.UserBase{
				Email: "test@test.com",
			},
		}, nil)

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	resp, err := handler(ctx, &struct{}{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "user-uuid-test", resp.Body.Id)
}

func Test_API_GetCurrentUser_NotFound(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetCurrentUser(svc.Service)

	svc.UserService.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, nil)

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	resp, err := handler(ctx, &struct{}{})

	require.Error(t, err)
	require.Nil(t, resp)
}

func Test_API_CreateUser_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateUser(svc.Service)

	input := &model.UserRequestBody{
		Body: model.UserCreate{
			UserBase: model.UserBase{
				Email:     "test@test.com",
				FirstName: "firstname",
				LastName:  "lastname",
			},
			Password: "secret",
		},
	}

	svc.UserService.
		EXPECT().
		Create(gomock.Any(), false).
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

func Test_API_CreateUser_AdminToken_DetectedAsAdmin(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateUser(svc.Service)

	input := &model.UserRequestBody{
		Authorization: "Bearer " + makeTestBearerToken(t, "admin-uuid-test", model.RoleAdmin),
		Body: model.UserCreate{
			UserBase: model.UserBase{
				Email:     "new-admin@test.com",
				FirstName: "New",
				LastName:  "Admin",
				Role:      model.RoleAdmin,
			},
			Password: "secret",
		},
	}

	svc.UserService.
		EXPECT().
		Create(gomock.Any(), true).
		Return(&model.User{Id: "user-uuid-test", UserBase: model.UserBase{Role: model.RoleAdmin}}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, model.RoleAdmin, resp.Body.Role)
}

func Test_API_CreateUser_InvalidBearerToken_TreatedAsNonAdmin(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateUser(svc.Service)

	input := &model.UserRequestBody{
		Authorization: "Bearer not-a-real-token",
	}

	svc.UserService.
		EXPECT().
		Create(gomock.Any(), false).
		Return(&model.User{Id: "user-uuid-test"}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_API_CreateUser_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateUser(svc.Service)

	svc.UserService.
		EXPECT().
		Create(gomock.Any(), false).
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

func Test_API_UpdateUser_Success_SelfUpdate(t *testing.T) {
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
		Update("user-uuid-test", gomock.Any(), false).
		Return(&model.User{
			Id: "user-uuid-test",
			UserBase: model.UserBase{
				FirstName: "updated-firstname",
			},
		}, nil)

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	resp, err := handler(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "updated-firstname", resp.Body.FirstName)
}

func Test_API_UpdateUser_Success_AdminUpdatesSomeoneElse(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateUser(svc.Service)

	input := &model.UserFilterFilterAndBody{
		UserRequestFilter: model.UserRequestFilter{UserId: "someone-else-uuid-test"},
		Body:              model.UserBase{Role: model.RoleAdmin},
	}

	svc.UserService.
		EXPECT().
		Update("someone-else-uuid-test", gomock.Any(), true).
		Return(&model.User{Id: "someone-else-uuid-test", UserBase: model.UserBase{Role: model.RoleAdmin}}, nil)

	ctx := context.WithValue(context.Background(), userIdContextKey, "admin-uuid-test")
	ctx = context.WithValue(ctx, roleContextKey, model.RoleAdmin)
	resp, err := handler(ctx, input)

	require.NoError(t, err)
	require.Equal(t, model.RoleAdmin, resp.Body.Role)
}

func Test_API_UpdateUser_Forbidden_NotSelfNotAdmin(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateUser(svc.Service)

	input := &model.UserFilterFilterAndBody{
		UserRequestFilter: model.UserRequestFilter{UserId: "someone-else-uuid-test"},
	}

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	resp, err := handler(ctx, input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "own profile")
}

func Test_API_UpdateUser_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateUser(svc.Service)

	svc.UserService.
		EXPECT().
		Update("user-uuid-test", gomock.Any(), false).
		Return(nil, errors.New("failed to update user"))

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	resp, err := handler(ctx, &model.UserFilterFilterAndBody{
		UserRequestFilter: model.UserRequestFilter{UserId: "user-uuid-test"},
	})

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

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	resp, err := handler(ctx, &model.UserPatchPasswordFilterAndBody{
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

func Test_API_PatchUserPassword_Forbidden_NotOwnAccount(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := PatchUserPassword(svc.Service)

	ctx := context.WithValue(context.Background(), userIdContextKey, "someone-else-uuid-test")
	resp, err := handler(ctx, &model.UserPatchPasswordFilterAndBody{
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
	require.ErrorContains(t, err, "own password")
}

func Test_API_PatchUserPassword_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := PatchUserPassword(svc.Service)

	svc.UserService.
		EXPECT().
		PatchPassword("user-uuid-test", gomock.Any()).
		Return(errors.New("failed to patch user password"))

	ctx := context.WithValue(context.Background(), userIdContextKey, "user-uuid-test")
	resp, err := handler(ctx, &model.UserPatchPasswordFilterAndBody{
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

func Test_API_ListUsers_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := ListUsers(svc.Service)

	svc.UserService.
		EXPECT().
		List().
		Return([]model.User{
			{Id: "user-uuid-test", UserBase: model.UserBase{Email: "test@test.com"}},
		}, nil)

	resp, err := handler(context.Background(), &struct{}{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Body, 1)
	require.Equal(t, "user-uuid-test", resp.Body[0].Id)
}

func Test_API_ListUsers_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := ListUsers(svc.Service)

	svc.UserService.
		EXPECT().
		List().
		Return(nil, errors.New("failed to list users"))

	resp, err := handler(context.Background(), &struct{}{})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to list users")
}
