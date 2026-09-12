package controller

import (
	"backend/internal/config"
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

		return mapper.MapSettingToSettingPageResponse(input.Body.LanguageCode, settings, svc.AuthService.IsOidcEnabled(), config.Bool("authentication.registrationEnabled"), config.Bool("authentication.emailVerification.enabled")), nil
	}
}

// UpdateSettingsBatch saves many settings in a single request. Invalid items
// never reach the database and valid items are saved even if others in the
// same batch fail - the response's `failures` list reports which keys were
// rejected and why, alongside the resulting (fully up to date) settings page.
func UpdateSettingsBatch(svc *domain.Service) func(c context.Context, input *model.SettingBatchRequest) (*model.SettingBatchResponse, error) {
	return func(c context.Context, input *model.SettingBatchRequest) (*model.SettingBatchResponse, error) {
		if len(input.Body.Settings) == 0 {
			return nil, huma.Error400BadRequest("settings must not be empty")
		}

		failures := svc.SettingService.UpdateMany(input.Body.Settings)

		settings, err := svc.SettingService.List()
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get settings", err)
		}

		// The batch's items all belong to the same page language in practice
		// (the frontend saves one language at a time) - use the first item's
		// to decide which language's page to render back.
		languageCode := input.Body.Settings[0].LanguageCode

		page := mapper.MapSettingToSettingPageResponse(languageCode, settings, svc.AuthService.IsOidcEnabled(), config.Bool("authentication.registrationEnabled"), config.Bool("authentication.emailVerification.enabled"))

		return &model.SettingBatchResponse{
			Body: model.SettingBatchResponseBody{
				Settings: page.Body,
				Failures: failures,
			},
		}, nil
	}
}

func GetPageSettings(svc *domain.Service) func(c context.Context, input *model.SettingRequestFiler) (*model.SettingPageResponse, error) {
	return func(c context.Context, input *model.SettingRequestFiler) (*model.SettingPageResponse, error) {
		settings, err := svc.SettingService.List()
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get settings", err)
		}

		return mapper.MapSettingToSettingPageResponse(input.LanguageCode, settings, svc.AuthService.IsOidcEnabled(), config.Bool("authentication.registrationEnabled"), config.Bool("authentication.emailVerification.enabled")), nil
	}
}

// GetEmailDeliveryInfo is admin-only and deliberately minimal (host + from
// address only) - SMTP itself is configured exclusively via config, never
// through the UI, so there's nothing here to edit.
func GetEmailDeliveryInfo(svc *domain.Service) func(c context.Context, input *struct{}) (*model.EmailDeliveryInfoResponse, error) {
	return func(c context.Context, input *struct{}) (*model.EmailDeliveryInfoResponse, error) {
		return &model.EmailDeliveryInfoResponse{
			Body: model.EmailDeliveryInfo{
				Enabled: config.Bool("authentication.emailVerification.enabled"),
				Host:    config.String("smtp.host"),
				From:    config.String("smtp.from"),
			},
		}, nil
	}
}
