package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"

	"github.com/danielgtaylor/huma/v2"
)

func CreateLink(svc *domain.Service) func(c context.Context, input *model.LinkRequestBody) (*model.LinkResponse, error) {
	return func(c context.Context, input *model.LinkRequestBody) (*model.LinkResponse, error) {
		link, err := svc.LinkService.Create(mapper.MapLinkBaseToLinkPointer(input.Body))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to create link", err)
		}

		return mapper.MapLinkToLinkResponse(*link), nil
	}
}

func GetLinks(svc *domain.Service) func(c context.Context, input *model.LinkRequestFilter) (*model.LinkResponseList, error) {
	return func(c context.Context, input *model.LinkRequestFilter) (*model.LinkResponseList, error) {
		links, err := svc.LinkService.List(input.ShelfId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get links", err)
		}

		return mapper.MapLinksToLinkResponseList(links), nil
	}
}

func UpdateLink(svc *domain.Service) func(c context.Context, input *model.LinkFilterFilterAndBody) (*model.LinkResponse, error) {
	return func(c context.Context, input *model.LinkFilterFilterAndBody) (*model.LinkResponse, error) {
		link, err := svc.LinkService.Update(input.LinkId, mapper.MapLinkBaseToLinkPointer(input.Body))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to update link", err)
		}

		return mapper.MapLinkToLinkResponse(*link), nil
	}
}

func DeleteLink(svc *domain.Service) func(c context.Context, input *model.LinkRequestFilter) (*struct{}, error) {
	return func(c context.Context, input *model.LinkRequestFilter) (*struct{}, error) {
		err := svc.LinkService.Delete(input.LinkId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to delete link", err)
		}
		return nil, nil
	}
}
