package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"

	"github.com/danielgtaylor/huma/v2"
)

func CreateUser(svc *domain.Service) func(c context.Context, input *model.UserRequestBody) (*model.UserResponse, error) {
	return func(c context.Context, input *model.UserRequestBody) (*model.UserResponse, error) {
		user, err := svc.UserService.Create(mapper.MapUserBaseToUserPointer(input.Body))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to create user", err)
		}

		return mapper.MapUserToUserResponse(*user), nil
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

func UpdateUser(svc *domain.Service) func(c context.Context, input *model.UserFilterFilterAndBody) (*model.UserResponse, error) {
	return func(c context.Context, input *model.UserFilterFilterAndBody) (*model.UserResponse, error) {
		user, err := svc.UserService.Update(input.UserId, mapper.MapUserBaseToUserPointer(input.Body))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to update user", err)
		}

		return mapper.MapUserToUserResponse(*user), nil
	}
}

func PatchUserPassword(svc *domain.Service) func(c context.Context, input *model.UserPatchPasswordFilterAndBody) (*struct{}, error) {
	return func(c context.Context, input *model.UserPatchPasswordFilterAndBody) (*struct{}, error) {
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
