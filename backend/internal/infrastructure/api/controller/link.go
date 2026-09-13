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
		link, err := svc.LinkService.Create(UserIdFromContext(c), IsAdminFromContext(c), mapper.MapLinkBaseToLinkPointer(input.Body))
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to create link", err)
		}

		return mapper.MapLinkToLinkResponse(*link), nil
	}
}

// GetLinks renders a shelf's links for its public link page and requires no
// authentication - it's deliberately unscoped by ownership.
func GetLinks(svc *domain.Service) func(c context.Context, input *model.LinkRequestShelfFilter) (*model.LinkResponseList, error) {
	return func(c context.Context, input *model.LinkRequestShelfFilter) (*model.LinkResponseList, error) {
		links, err := svc.LinkService.List(input.ShelfId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get links", err)
		}

		return mapper.MapLinksToLinkResponseList(links), nil
	}
}

func UpdateLink(svc *domain.Service) func(c context.Context, input *model.LinkFilterFilterAndBody) (*model.LinkResponse, error) {
	return func(c context.Context, input *model.LinkFilterFilterAndBody) (*model.LinkResponse, error) {
		link, err := svc.LinkService.Update(input.LinkId, UserIdFromContext(c), IsAdminFromContext(c), mapper.MapLinkBaseToLinkPointer(input.Body))
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to update link", err)
		}

		return mapper.MapLinkToLinkResponse(*link), nil
	}
}

func DeleteLink(svc *domain.Service) func(c context.Context, input *model.LinkRequestFilter) (*struct{}, error) {
	return func(c context.Context, input *model.LinkRequestFilter) (*struct{}, error) {
		err := svc.LinkService.Delete(input.LinkId, UserIdFromContext(c), IsAdminFromContext(c))
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to delete link", err)
		}
		return nil, nil
	}
}

// UpdateLinksOrder saves the order of many links in a single request. Invalid
// items (unknown id, or one the caller doesn't own) never abort the rest of
// the batch - they're reported back in the response's failures list. A link
// can only be reordered within the section it already belongs to.
func UpdateLinksOrder(svc *domain.Service) func(c context.Context, input *model.LinkOrderRequest) (*model.LinkOrderResponse, error) {
	return func(c context.Context, input *model.LinkOrderRequest) (*model.LinkOrderResponse, error) {
		if len(input.Body.Links) == 0 {
			return nil, huma.Error400BadRequest("links must not be empty")
		}

		failures := svc.LinkService.UpdateOrder(UserIdFromContext(c), IsAdminFromContext(c), input.Body.Links)

		return &model.LinkOrderResponse{
			Body: model.LinkOrderResponseBody{Failures: failures},
		}, nil
	}
}
