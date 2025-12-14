package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"

	"github.com/danielgtaylor/huma/v2"
)

func CreateShelf(svc *domain.Service) func(c context.Context, input *model.ShelfRequestBody) (*model.ShelfResponse, error) {
	return func(c context.Context, input *model.ShelfRequestBody) (*model.ShelfResponse, error) {
		userId, err := svc.ShelfService.CreateShelf(mapper.MapShelfBaseToShelfPointer(input.Body))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to create user", err)
		}

		user, err := svc.ShelfService.GetShelfById(userId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get user", err)
		}

		return mapper.MapShelfToShelfResponse(*user), nil
	}
}

func GetShelfById(svc *domain.Service) func(c context.Context, input *model.ShelfRequestFilter) (*model.ShelfResponse, error) {
	return func(c context.Context, input *model.ShelfRequestFilter) (*model.ShelfResponse, error) {
		user, err := svc.ShelfService.GetShelfById(input.ShelfId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get user", err)
		}

		return mapper.MapShelfToShelfResponse(*user), nil
	}
}

func UpdateShelf(svc *domain.Service) func(c context.Context, input *model.ShelfFilterFilterAndBody) (*model.ShelfResponse, error) {
	return func(c context.Context, input *model.ShelfFilterFilterAndBody) (*model.ShelfResponse, error) {
		shelf, err := svc.ShelfService.UpdateShelf(input.ShelfId, mapper.MapShelfBaseToShelfPointer(input.Body))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to update user", err)
		}

		return mapper.MapShelfToShelfResponse(*shelf), nil
	}
}

func DeleteShelf(svc *domain.Service) func(c context.Context, input *model.ShelfRequestFilter) (*struct{}, error) {
	return func(c context.Context, input *model.ShelfRequestFilter) (*struct{}, error) {
		user, err := svc.ShelfService.GetShelfById(input.ShelfId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get user", err)
		}

		err = svc.ShelfService.DeleteShelf(user)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to delete user", err)
		}

		return nil, nil
	}
}
