package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
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

// GetUserLimits returns a user's shelf limit and usage. Admin or service token.
func GetUserLimits(svc *domain.Service) func(c context.Context, input *model.UserRequestFilter) (*model.UserLimitsResponse, error) {
	return func(c context.Context, input *model.UserRequestFilter) (*model.UserLimitsResponse, error) {
		limits, err := svc.UserService.GetLimits(input.UserId)
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to get user limits", err)
		}
		return &model.UserLimitsResponse{Body: *limits}, nil
	}
}

// PatchUserLimits sets a user's shelf limit. Admin or service token.
func PatchUserLimits(svc *domain.Service) func(c context.Context, input *model.UserLimitsPatchRequest) (*model.UserLimitsResponse, error) {
	return func(c context.Context, input *model.UserLimitsPatchRequest) (*model.UserLimitsResponse, error) {
		if input.Body.MaxShelves != nil && *input.Body.MaxShelves < 0 {
			return nil, huma.Error400BadRequest("validation failed: max_shelves must be zero or a positive number")
		}

		limits, previous, err := svc.UserService.SetMaxShelves(input.UserId, input.Body.MaxShelves)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidInput) {
				return nil, huma.Error400BadRequest("validation failed: max_shelves must be zero or a positive number")
			}
			return nil, mapper.MapOwnershipError("failed to set user limits", err)
		}

		if IsServiceTokenFromContext(c) {
			zap.L().Info("service token changed user limit",
				zap.String("user_id", input.UserId),
				zap.String("old_max_shelves", formatLimit(previous)),
				zap.String("new_max_shelves", formatLimit(limits.MaxShelves)))
		}
		return &model.UserLimitsResponse{Body: *limits}, nil
	}
}

func formatLimit(n *int) string {
	if n == nil {
		return "unlimited"
	}
	return strconv.Itoa(*n)
}
