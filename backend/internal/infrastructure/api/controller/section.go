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
		section, err := svc.SectionService.Create(UserIdFromContext(c), IsAdminFromContext(c), mapper.MapSectionBaseToSectionPointer(input.Body))
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to create section", err)
		}

		return mapper.MapSectionToSectionResponse(*section), nil
	}
}

// GetSections renders a shelf's sections for its public link page and
// requires no authentication - it's deliberately unscoped by ownership.
func GetSections(svc *domain.Service) func(c context.Context, input *model.SectionRequestShelfFilter) (*model.SectionResponseList, error) {
	return func(c context.Context, input *model.SectionRequestShelfFilter) (*model.SectionResponseList, error) {
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

func UpdateSection(svc *domain.Service) func(c context.Context, input *model.SectionRequestSectionFilterAndBody) (*model.SectionResponse, error) {
	return func(c context.Context, input *model.SectionRequestSectionFilterAndBody) (*model.SectionResponse, error) {
		if strings.TrimSpace(input.SectionId) == "" {
			return nil, huma.Error400BadRequest("sectionId is required", nil)
		}

		section, err := svc.SectionService.Update(input.SectionId, UserIdFromContext(c), IsAdminFromContext(c), mapper.MapSectionBaseToSectionPointer(input.Body))
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to update section", err)
		}

		return mapper.MapSectionToSectionResponse(*section), nil
	}
}

func DeleteSection(svc *domain.Service) func(c context.Context, input *model.SectionRequestSectionFilter) (*struct{}, error) {
	return func(c context.Context, input *model.SectionRequestSectionFilter) (*struct{}, error) {
		err := svc.SectionService.Delete(input.SectionId, UserIdFromContext(c), IsAdminFromContext(c))
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to delete section", err)
		}

		return nil, nil
	}
}
