//go:generate mockgen -source=setting.go -destination=mocks/setting_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"
)

// settingKeys are the keys the setting page reads.
var settingKeys = map[string]bool{
	"about":                 true,
	"about_show":            true,
	"contact":               true,
	"contact_show":          true,
	"imprint":               true,
	"imprint_show":          true,
	"terms_of_use":          true,
	"terms_of_use_show":     true,
	"privacy_policy":        true,
	"privacy_policy_show":   true,
	"redirect_to_dashboard": true,
}

// booleanSettingKeys must be "true" or "false".
var booleanSettingKeys = map[string]bool{
	"about_show":            true,
	"contact_show":          true,
	"imprint_show":          true,
	"terms_of_use_show":     true,
	"privacy_policy_show":   true,
	"redirect_to_dashboard": true,
}

// supportedLanguageCodes mirrors i18n.locales in frontend/nuxt.config.ts.
var supportedLanguageCodes = map[string]bool{
	"en":    true,
	"de":    true,
	"de-CH": true,
	"es":    true,
}

type SettingService interface {
	List() ([]model.Setting, error)
	Update(setting model.Setting) error
	// UpdateMany saves valid settings and reports the rest as failures.
	UpdateMany(settings []model.Setting) []model.SettingUpdateFailure
}

type settingServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
}

func NewSettingService(repository *repository.Repository, domain *Service) SettingService {
	return &settingServiceImpl{
		Repository: repository,
		Domain:     domain,
	}
}

func (s *settingServiceImpl) List() ([]model.Setting, error) {
	return s.Repository.SettingRepository.List()
}

func (s *settingServiceImpl) Update(setting model.Setting) error {
	return s.Repository.SettingRepository.Upsert(setting.Key, setting.LanguageCode, setting.Value)
}

func (s *settingServiceImpl) UpdateMany(settings []model.Setting) []model.SettingUpdateFailure {
	var failures []model.SettingUpdateFailure

	for _, setting := range settings {
		if reason := validateSetting(setting); reason != "" {
			failures = append(failures, model.SettingUpdateFailure{
				Key:          setting.Key,
				LanguageCode: setting.LanguageCode,
				Reason:       reason,
			})
			continue
		}

		if err := s.Repository.SettingRepository.Upsert(setting.Key, setting.LanguageCode, setting.Value); err != nil {
			failures = append(failures, model.SettingUpdateFailure{
				Key:          setting.Key,
				LanguageCode: setting.LanguageCode,
				Reason:       err.Error(),
			})
		}
	}

	return failures
}

// validateSetting returns why a setting is invalid, or "".
func validateSetting(setting model.Setting) string {
	if !settingKeys[setting.Key] {
		return fmt.Sprintf("unknown setting key %q", setting.Key)
	}
	if !supportedLanguageCodes[setting.LanguageCode] {
		return fmt.Sprintf("unsupported language code %q", setting.LanguageCode)
	}
	if booleanSettingKeys[setting.Key] && setting.Value != "true" && setting.Value != "false" {
		return fmt.Sprintf("value for %q must be \"true\" or \"false\"", setting.Key)
	}
	return ""
}
