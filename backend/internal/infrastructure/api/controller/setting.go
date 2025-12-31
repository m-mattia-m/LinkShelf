package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"

	"github.com/danielgtaylor/huma/v2"
)

func UpdateSetting(svc *domain.Service) func(c context.Context, input *model.SettingRequest) (*model.SettingPageResponse, error) {
	return func(c context.Context, input *model.SettingRequest) (*model.SettingPageResponse, error) {
		err := svc.SettingService.Update(input.Body)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to update setting", err)
		}

		settings, err := svc.SettingService.List()
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get setting", err)
		}

		return mapper.MapSettingToSettingPageResponse(input.Body.LanguageCode, settings), nil
	}
}

func GetPageSettings(svc *domain.Service) func(c context.Context, input *model.SettingRequestFiler) (*model.SettingPageResponse, error) {
	return func(c context.Context, input *model.SettingRequestFiler) (*model.SettingPageResponse, error) {
		settings, err := svc.SettingService.List()
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get settings", err)
		}

		return mapper.MapSettingToSettingPageResponse(input.LanguageCode, settings), nil
	}
}
