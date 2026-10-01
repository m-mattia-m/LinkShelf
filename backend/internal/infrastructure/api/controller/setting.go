package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/mailer"
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

		return mapper.MapSettingToSettingPageResponse(input.Body.LanguageCode, settings, settingPageFlags(svc)), nil
	}
}

// UpdateSettingsBatch saves valid items and reports the rest as failures.
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

		// The frontend saves one language at a time.
		languageCode := input.Body.Settings[0].LanguageCode

		page := mapper.MapSettingToSettingPageResponse(languageCode, settings, settingPageFlags(svc))

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

		return mapper.MapSettingToSettingPageResponse(input.LanguageCode, settings, settingPageFlags(svc)), nil
	}
}

// GetEmailDeliveryInfo returns the SMTP host and from address. Admin-only.
func GetEmailDeliveryInfo(svc *domain.Service) func(c context.Context, input *struct{}) (*model.EmailDeliveryInfoResponse, error) {
	return func(c context.Context, input *struct{}) (*model.EmailDeliveryInfoResponse, error) {
		return &model.EmailDeliveryInfoResponse{
			Body: model.EmailDeliveryInfo{
				Enabled: config.Bool("authentication.emailVerification.enabled"),
				Host:    config.String("smtp.host"),
				From:    mailer.FormatFromHeader(config.String("smtp.from"), config.String("smtp.fromName")),
			},
		}, nil
	}
}

// settingPageFlags returns the config-derived values of the settings page.
func settingPageFlags(svc *domain.Service) mapper.SettingPageFlags {
	return mapper.SettingPageFlags{
		OidcEnabled: svc.AuthService.IsOidcEnabled(),
		// Sign-up and reset depend on local auth.
		RegistrationEnabled:      config.Bool("authentication.registrationEnabled") && config.LocalAuthEnabled(),
		EmailVerificationEnabled: config.Bool("authentication.emailVerification.enabled"),
		UserBasedPaths:           config.Bool("app.userBasedPaths"),
		PasswordResetEnabled:     config.Bool("authentication.passwordReset.enabled") && config.LocalAuthEnabled(),
		LocalAuthEnabled:         config.LocalAuthEnabled(),
	}
}
