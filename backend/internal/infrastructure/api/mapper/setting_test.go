package mapper

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MapSettingToSettingPageResponse_Success(t *testing.T) {
	settings := []model.Setting{
		{Key: "about_show", LanguageCode: "en", Value: "true"},
		{Key: "about", LanguageCode: "en", Value: "About EN"},
		{Key: "contact_show", LanguageCode: "en", Value: "false"},
		{Key: "contact", LanguageCode: "en", Value: "Contact EN"},
		{Key: "imprint_show", LanguageCode: "en", Value: "true"},
		{Key: "imprint", LanguageCode: "en", Value: "Imprint EN"},
		{Key: "terms_of_use_show", LanguageCode: "en", Value: "true"},
		{Key: "terms_of_use", LanguageCode: "en", Value: "Terms EN"},
		{Key: "privacy_policy_show", LanguageCode: "en", Value: "false"},
		{Key: "privacy_policy", LanguageCode: "en", Value: "Privacy EN"},
	}

	resp := MapSettingToSettingPageResponse("en", settings)

	require.NotNil(t, resp)

	require.True(t, resp.Body.AboutShow)
	require.Equal(t, "About EN", resp.Body.About)

	require.False(t, resp.Body.ContactShow)
	require.Equal(t, "Contact EN", resp.Body.Contact)

	require.True(t, resp.Body.ImprintShow)
	require.Equal(t, "Imprint EN", resp.Body.Imprint)

	require.True(t, resp.Body.TermsOfUseShow)
	require.Equal(t, "Terms EN", resp.Body.TermsOfUse)

	require.False(t, resp.Body.PrivacyPolicyShow)
	require.Equal(t, "Privacy EN", resp.Body.PrivacyPolicy)
}

func Test_MapSettingToSettingPageResponse_LanguageIsolation(t *testing.T) {
	settings := []model.Setting{
		{Key: "about_show", LanguageCode: "en", Value: "true"},
		{Key: "about", LanguageCode: "en", Value: "About EN"},
		{Key: "about_show", LanguageCode: "de", Value: "false"},
		{Key: "about", LanguageCode: "de", Value: "Über DE"},
	}

	resp := MapSettingToSettingPageResponse("de", settings)

	require.False(t, resp.Body.AboutShow)
	require.Equal(t, "Über DE", resp.Body.About)
}

func Test_MapSettingToSettingPageResponse_BooleanParsing(t *testing.T) {
	settings := []model.Setting{
		{Key: "about_show", LanguageCode: "en", Value: "TRUE"},
		{Key: "about", LanguageCode: "en", Value: "About"},
	}

	resp := MapSettingToSettingPageResponse("en", settings)

	// Only exact "true" is treated as true
	require.False(t, resp.Body.AboutShow)
}

func Test_MapSettingToSettingPageResponse_MissingSettings(t *testing.T) {
	resp := MapSettingToSettingPageResponse("en", nil)

	require.NotNil(t, resp)

	require.False(t, resp.Body.AboutShow)
	require.Empty(t, resp.Body.About)

	require.False(t, resp.Body.ContactShow)
	require.Empty(t, resp.Body.Contact)

	require.False(t, resp.Body.ImprintShow)
	require.Empty(t, resp.Body.Imprint)

	require.False(t, resp.Body.TermsOfUseShow)
	require.Empty(t, resp.Body.TermsOfUse)

	require.False(t, resp.Body.PrivacyPolicyShow)
	require.Empty(t, resp.Body.PrivacyPolicy)
}

func Test_MapSettingToSettingPageResponse_LoginOptions(t *testing.T) {
	settings := []model.Setting{
		{Key: "login_options", LanguageCode: "en", Value: `["Local", "Zitadel"]`},
	}

	resp := MapSettingToSettingPageResponse("en", settings)

	require.Equal(t, []string{"Local", "Zitadel"}, resp.Body.LoginOptions)
}

func Test_MapSettingToSettingPageResponse_LoginOptions_Missing(t *testing.T) {
	resp := MapSettingToSettingPageResponse("en", nil)

	require.Nil(t, resp.Body.LoginOptions)
}

func Test_MapSettingToSettingPageResponse_LoginOptions_Invalid(t *testing.T) {
	settings := []model.Setting{
		{Key: "login_options", LanguageCode: "en", Value: `not-json`},
	}

	resp := MapSettingToSettingPageResponse("en", settings)

	require.Nil(t, resp.Body.LoginOptions)
}
