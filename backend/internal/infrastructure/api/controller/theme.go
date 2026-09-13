package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"

	"github.com/danielgtaylor/huma/v2"
)

// ListThemes returns instance-provided themes alongside the caller's own,
// grouped the way the shelf editor's theme picker and the "Themes" nav page
// render them.
func ListThemes(svc *domain.Service) func(c context.Context, input *struct{}) (*model.ThemeGroupedResponse, error) {
	return func(c context.Context, input *struct{}) (*model.ThemeGroupedResponse, error) {
		grouped, err := svc.ThemeService.ListGrouped(UserIdFromContext(c))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to list themes", err)
		}
		return mapper.MapThemeGroupedToResponse(grouped), nil
	}
}

// ListAllUserThemes is the admin moderation view across every user's themes.
func ListAllUserThemes(svc *domain.Service) func(c context.Context, input *struct{}) (*model.ThemeListResponse, error) {
	return func(c context.Context, input *struct{}) (*model.ThemeListResponse, error) {
		themes, err := svc.ThemeService.ListAllUserScoped()
		if err != nil {
			return nil, huma.Error400BadRequest("failed to list themes", err)
		}
		return mapper.MapThemesToThemeListResponse(themes), nil
	}
}

func GetTheme(svc *domain.Service) func(c context.Context, input *model.ThemeRequestFilter) (*model.ThemeResponse, error) {
	return func(c context.Context, input *model.ThemeRequestFilter) (*model.ThemeResponse, error) {
		theme, err := svc.ThemeService.Get(input.ThemeId, UserIdFromContext(c), IsAdminFromContext(c))
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to get theme", err)
		}
		return mapper.MapThemeToThemeResponse(*theme), nil
	}
}

func CreateTheme(svc *domain.Service) func(c context.Context, input *model.ThemeRequestBody) (*model.ThemeResponse, error) {
	return func(c context.Context, input *model.ThemeRequestBody) (*model.ThemeResponse, error) {
		theme, err := svc.ThemeService.Create(UserIdFromContext(c), input.Body)
		if err != nil {
			return nil, mapper.MapWriteError("failed to create theme", err)
		}
		return mapper.MapThemeToThemeResponse(*theme), nil
	}
}

func UpdateTheme(svc *domain.Service) func(c context.Context, input *model.ThemeFilterAndBody) (*model.ThemeResponse, error) {
	return func(c context.Context, input *model.ThemeFilterAndBody) (*model.ThemeResponse, error) {
		theme, err := svc.ThemeService.Update(input.ThemeId, UserIdFromContext(c), input.Body)
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to update theme", err)
		}
		return mapper.MapThemeToThemeResponse(*theme), nil
	}
}

func DeleteTheme(svc *domain.Service) func(c context.Context, input *model.ThemeRequestFilter) (*struct{}, error) {
	return func(c context.Context, input *model.ThemeRequestFilter) (*struct{}, error) {
		err := svc.ThemeService.Delete(input.ThemeId, UserIdFromContext(c), IsAdminFromContext(c))
		if err != nil {
			return nil, mapper.MapOwnershipError("failed to delete theme", err)
		}
		return nil, nil
	}
}
