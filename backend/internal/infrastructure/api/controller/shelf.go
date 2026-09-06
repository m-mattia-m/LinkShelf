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
		shelfId, err := svc.ShelfService.Create(mapper.MapShelfBaseToShelfPointer(input.Body))
		if err != nil {
			return nil, mapper.MapWriteError("failed to create shelf", err)
		}

		shelf, err := svc.ShelfService.Get(shelfId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get shelf", err)
		}

		return mapper.MapShelfToShelfResponse(*shelf), nil
	}
}

func ListShelf(svc *domain.Service) func(c context.Context, input *struct{}) (*model.ShelfListResponse, error) {
	return func(c context.Context, input *struct{}) (*model.ShelfListResponse, error) {
		shelves, err := svc.ShelfService.List()
		if err != nil {
			return nil, huma.Error400BadRequest("failed to list shelves", err)
		}

		return mapper.MapShelfToShelfListResponse(shelves), nil
	}
}

func GetShelfById(svc *domain.Service) func(c context.Context, input *model.ShelfRequestFilter) (*model.ShelfResponse, error) {
	return func(c context.Context, input *model.ShelfRequestFilter) (*model.ShelfResponse, error) {
		shelf, err := svc.ShelfService.Get(input.ShelfId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get shelf", err)
		}

		return mapper.MapShelfToShelfResponse(*shelf), nil
	}
}

func GetPublicShelfByPath(svc *domain.Service) func(c context.Context, input *model.ShelfPathFilter) (*model.PublicShelfResponse, error) {
	return func(c context.Context, input *model.ShelfPathFilter) (*model.PublicShelfResponse, error) {
		shelf, err := svc.ShelfService.GetByPath(input.Path)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get shelf", err)
		}

		if shelf == nil {
			return nil, huma.Error404NotFound("shelf not found")
		}

		return mapper.MapShelfToPublicShelfResponse(*shelf), nil
	}
}

func UpdateShelf(svc *domain.Service) func(c context.Context, input *model.ShelfFilterFilterAndBody) (*model.ShelfResponse, error) {
	return func(c context.Context, input *model.ShelfFilterFilterAndBody) (*model.ShelfResponse, error) {
		shelf, err := svc.ShelfService.Update(input.ShelfId, mapper.MapShelfBaseToShelfPointer(input.Body))
		if err != nil {
			return nil, mapper.MapWriteError("failed to update shelf", err)
		}

		return mapper.MapShelfToShelfResponse(*shelf), nil
	}
}

func DeleteShelf(svc *domain.Service) func(c context.Context, input *model.ShelfRequestFilter) (*struct{}, error) {
	return func(c context.Context, input *model.ShelfRequestFilter) (*struct{}, error) {
		shelf, err := svc.ShelfService.Get(input.ShelfId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get shelf", err)
		}

		err = svc.ShelfService.Delete(shelf)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to delete shelf", err)
		}

		return nil, nil
	}
}
