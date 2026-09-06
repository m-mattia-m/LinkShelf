package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// CreateUser is public (self-registration), but also doubles as the
// admin-facing "add a user" endpoint: when called with a valid admin Bearer
// token, the request may also set the new user's role. Since the operation
// carries no Security requirement, the middleware never validates this
// token for us - it's checked manually here, same pattern as OidcCallback.
func CreateUser(svc *domain.Service) func(c context.Context, input *model.UserRequestBody) (*model.UserResponse, error) {
	return func(c context.Context, input *model.UserRequestBody) (*model.UserResponse, error) {
		isAdmin := false
		if bearer := strings.TrimSpace(strings.TrimPrefix(input.Authorization, "Bearer ")); bearer != "" {
			if claims, err := domain.ValidateAccessToken(bearer); err == nil {
				isAdmin = claims.Role == model.RoleAdmin
			}
		}

		user, err := svc.UserService.Create(&input.Body, isAdmin)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidRole) {
				return nil, huma.Error400BadRequest(err.Error())
			}
			return nil, mapper.MapWriteError("failed to create user", err)
		}

		return mapper.MapUserToUserResponse(*user), nil
	}
}

func GetCurrentUser(svc *domain.Service) func(c context.Context, input *struct{}) (*model.UserResponse, error) {
	return func(c context.Context, input *struct{}) (*model.UserResponse, error) {
		user, err := svc.UserService.Get(UserIdFromContext(c))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get current user", err)
		}
		if user == nil {
			return nil, huma.Error404NotFound("user not found")
		}

		return mapper.MapUserToUserResponse(*user), nil
	}
}

func ListUsers(svc *domain.Service) func(c context.Context, input *struct{}) (*model.UserListResponse, error) {
	return func(c context.Context, input *struct{}) (*model.UserListResponse, error) {
		users, err := svc.UserService.List()
		if err != nil {
			return nil, huma.Error400BadRequest("failed to list users", err)
		}

		return mapper.MapUsersToUserListResponse(users), nil
	}
}

func GetUserById(svc *domain.Service) func(c context.Context, input *model.UserRequestFilter) (*model.UserResponse, error) {
	return func(c context.Context, input *model.UserRequestFilter) (*model.UserResponse, error) {
		user, err := svc.UserService.Get(input.UserId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get user", err)
		}

		return mapper.MapUserToUserResponse(*user), nil
	}
}

// UpdateUser lets a user update their own profile, or an admin update
// anyone's - but only an admin caller may change the role field.
func UpdateUser(svc *domain.Service) func(c context.Context, input *model.UserFilterFilterAndBody) (*model.UserResponse, error) {
	return func(c context.Context, input *model.UserFilterFilterAndBody) (*model.UserResponse, error) {
		isAdmin := IsAdminFromContext(c)
		if input.UserId != UserIdFromContext(c) && !isAdmin {
			return nil, huma.Error403Forbidden("you may only update your own profile")
		}

		user, err := svc.UserService.Update(input.UserId, mapper.MapUserBaseToUserPointer(input.Body), isAdmin)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidRole) {
				return nil, huma.Error400BadRequest(err.Error())
			}
			return nil, mapper.MapWriteError("failed to update user", err)
		}
		if user == nil {
			return nil, huma.Error404NotFound("user not found")
		}

		return mapper.MapUserToUserResponse(*user), nil
	}
}

// PatchUserPassword only ever lets someone patch their own password - not
// even an admin may patch another user's password through this endpoint,
// since it requires knowing the current password.
func PatchUserPassword(svc *domain.Service) func(c context.Context, input *model.UserPatchPasswordFilterAndBody) (*struct{}, error) {
	return func(c context.Context, input *model.UserPatchPasswordFilterAndBody) (*struct{}, error) {
		if input.UserId != UserIdFromContext(c) {
			return nil, huma.Error403Forbidden("you may only change your own password")
		}

		err := svc.UserService.PatchPassword(input.UserId, &input.Body)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to patch user password", err)
		}

		return nil, nil
	}
}

func DeleteUser(svc *domain.Service) func(c context.Context, input *model.UserRequestFilter) (*struct{}, error) {
	return func(c context.Context, input *model.UserRequestFilter) (*struct{}, error) {
		user, err := svc.UserService.Get(input.UserId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get user", err)
		}

		err = svc.UserService.Delete(user)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to delete user", err)
		}

		return nil, nil
	}
}
