package mapper

import (
	"backend/internal/infrastructure/api/model"
	"fmt"
)

func MapSettingToSettingPageResponse(languageCode string, settings []model.Setting) *model.SettingPageResponse {
	settingsMap := make(map[string]model.Setting)

	for _, setting := range settings {
		settingsMap[fmt.Sprintf("%s_%s", setting.Key, setting.LanguageCode)] = setting
	}

	return &model.SettingPageResponse{
		Body: model.SettingPageBody{
			AboutShow:           getSettingValue(settingsMap, "about_show", languageCode) == "true",
			About:               getSettingValue(settingsMap, "about", languageCode),
			ContactShow:         getSettingValue(settingsMap, "contact_show", languageCode) == "true",
			Contact:             getSettingValue(settingsMap, "contact", languageCode),
			ImprintShow:         getSettingValue(settingsMap, "imprint_show", languageCode) == "true",
			Imprint:             getSettingValue(settingsMap, "imprint", languageCode),
			TermsOfUseShow:      getSettingValue(settingsMap, "terms_of_use_show", languageCode) == "true",
			TermsOfUse:          getSettingValue(settingsMap, "terms_of_use", languageCode),
			PrivacyPolicyShow:   getSettingValue(settingsMap, "privacy_policy_show", languageCode) == "true",
			PrivacyPolicy:       getSettingValue(settingsMap, "privacy_policy", languageCode),
			RedirectToDashboard: getSettingValue(settingsMap, "redirect_to_dashboard", languageCode) == "true",
		},
	}
}

func getSettingValue(settingsMap map[string]model.Setting, key string, languageCode string) string {
	value, found := settingsMap[fmt.Sprintf("%s_%s", key, languageCode)]
	if !found {
		languageCode = "en" // Fallback to English
		value, found = settingsMap[fmt.Sprintf("%s_%s", key, languageCode)]
		if !found {
			return ""
		}
	}
	return value.Value
}
