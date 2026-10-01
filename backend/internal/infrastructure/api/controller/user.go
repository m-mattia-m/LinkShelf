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

// CreateUser handles self-registration and admin user creation (admins may set a role).
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
			if errors.Is(err, domain.ErrInvalidRole) || errors.Is(err, domain.ErrInvalidInput) {
				return nil, huma.Error400BadRequest(err.Error())
			}
			if errors.Is(err, domain.ErrConflict) {
				return nil, huma.Error409Conflict(err.Error())
			}
			if errors.Is(err, domain.ErrRegistrationDisabled) {
				return nil, huma.Error403Forbidden(err.Error())
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

// UpdateUser updates the caller's own profile, or any profile for admins. Only admins may change roles.
func UpdateUser(svc *domain.Service) func(c context.Context, input *model.UserFilterFilterAndBody) (*model.UserResponse, error) {
	return func(c context.Context, input *model.UserFilterFilterAndBody) (*model.UserResponse, error) {
		isAdmin := IsAdminFromContext(c)
		if input.UserId != UserIdFromContext(c) && !isAdmin {
			return nil, huma.Error403Forbidden("you may only update your own profile")
		}

		user, err := svc.UserService.Update(input.UserId, mapper.MapUserBaseToUserPointer(input.Body), isAdmin)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidRole) || errors.Is(err, domain.ErrInvalidInput) {
				return nil, huma.Error400BadRequest(err.Error())
			}
			if errors.Is(err, domain.ErrConflict) {
				return nil, huma.Error409Conflict(err.Error())
			}
			if errors.Is(err, domain.ErrTooManyRequests) {
				return nil, huma.Error429TooManyRequests(err.Error())
			}
			return nil, mapper.MapWriteError("failed to update user", err)
		}
		if user == nil {
			return nil, huma.Error404NotFound("user not found")
		}

		return mapper.MapUserToUserResponse(*user), nil
	}
}

// PatchUserPassword changes the caller's own password.
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

// MarkUserVerified lets an admin mark an email verified, e.g. when SMTP is broken.
func MarkUserVerified(svc *domain.Service) func(c context.Context, input *model.UserRequestFilter) (*model.UserResponse, error) {
	return func(c context.Context, input *model.UserRequestFilter) (*model.UserResponse, error) {
		if err := svc.EmailVerificationService.MarkVerified(input.UserId); err != nil {
			return nil, huma.Error400BadRequest("failed to mark user as verified", err)
		}

		user, err := svc.UserService.Get(input.UserId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get user", err)
		}
		if user == nil {
			return nil, huma.Error404NotFound("user not found")
		}

		return mapper.MapUserToUserResponse(*user), nil
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
