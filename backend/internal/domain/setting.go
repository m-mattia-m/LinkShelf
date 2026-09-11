//go:generate mockgen -source=setting.go -destination=mocks/setting_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"
)

// settingKeys are the only keys the setting page actually reads (see
// model.SettingPageBody / mapper.MapSettingToSettingPageResponse) - anything
// else would silently write an orphan row nobody ever reads back.
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

// booleanSettingKeys are the keys mapped to a bool (via `== "true"`) by
// mapper.MapSettingToSettingPageResponse - their value must actually be
// "true" or "false".
var booleanSettingKeys = map[string]bool{
	"about_show":            true,
	"contact_show":          true,
	"imprint_show":          true,
	"terms_of_use_show":     true,
	"privacy_policy_show":   true,
	"redirect_to_dashboard": true,
}

// supportedLanguageCodes mirrors frontend/nuxt.config.ts's i18n.locales -
// keep both lists in sync if a locale is ever added or removed.
var supportedLanguageCodes = map[string]bool{
	"en":    true,
	"de":    true,
	"de-CH": true,
}

type SettingService interface {
	List() ([]model.Setting, error)
	Update(setting model.Setting) error
	// UpdateMany saves every valid setting and reports the rest as failures -
	// it never aborts the whole batch over one bad item.
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

// validateSetting returns a human-readable reason the setting can't be
// saved, or "" if it's valid.
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
