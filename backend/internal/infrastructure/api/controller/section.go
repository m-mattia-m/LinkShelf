package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func CreateSection(svc *domain.Service) func(c context.Context, input *model.SectionRequestBody) (*model.SectionResponse, error) {
	return func(c context.Context, input *model.SectionRequestBody) (*model.SectionResponse, error) {
		section, err := svc.SectionService.Create(mapper.MapSectionBaseToSectionPointer(input.Body))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to create section", err)
		}

		return mapper.MapSectionToSectionResponse(*section), nil
	}
}

func GetSections(svc *domain.Service) func(c context.Context, input *model.SectionRequestFilter) (*model.SectionResponseList, error) {
	return func(c context.Context, input *model.SectionRequestFilter) (*model.SectionResponseList, error) {
		if strings.TrimSpace(input.ShelfId) == "" {
			return nil, huma.Error400BadRequest("shelfId is required", nil)
		}

		sections, err := svc.SectionService.List(input.ShelfId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get section", err)
		}

		return mapper.MapSectionsToSectionResponseList(sections), nil
	}
}

func UpdateSection(svc *domain.Service) func(c context.Context, input *model.SectionFilterFilterAndBody) (*model.SectionResponse, error) {
	return func(c context.Context, input *model.SectionFilterFilterAndBody) (*model.SectionResponse, error) {
		if strings.TrimSpace(input.ShelfId) == "" {
			return nil, huma.Error400BadRequest("shelfId is required", nil)
		}

		section, err := svc.SectionService.Update(input.SectionId, mapper.MapSectionBaseToSectionPointer(input.Body))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to update section", err)
		}

		return mapper.MapSectionToSectionResponse(*section), nil
	}
}

func DeleteSection(svc *domain.Service) func(c context.Context, input *model.SectionRequestFilter) (*struct{}, error) {
	return func(c context.Context, input *model.SectionRequestFilter) (*struct{}, error) {
		err := svc.SectionService.Delete(input.SectionId)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to delete section", err)
		}

		return nil, nil
	}
}
