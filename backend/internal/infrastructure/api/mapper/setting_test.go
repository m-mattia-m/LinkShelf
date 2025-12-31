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

	require.True(t, resp.AboutShow)
	require.Equal(t, "About EN", resp.About)

	require.False(t, resp.ContactShow)
	require.Equal(t, "Contact EN", resp.Contact)

	require.True(t, resp.ImprintShow)
	require.Equal(t, "Imprint EN", resp.Imprint)

	require.True(t, resp.TermsOfUseShow)
	require.Equal(t, "Terms EN", resp.TermsOfUse)

	require.False(t, resp.PrivacyPolicyShow)
	require.Equal(t, "Privacy EN", resp.PrivacyPolicy)
}

func Test_MapSettingToSettingPageResponse_LanguageIsolation(t *testing.T) {
	settings := []model.Setting{
		{Key: "about_show", LanguageCode: "en", Value: "true"},
		{Key: "about", LanguageCode: "en", Value: "About EN"},
		{Key: "about_show", LanguageCode: "de", Value: "false"},
		{Key: "about", LanguageCode: "de", Value: "Über DE"},
	}

	resp := MapSettingToSettingPageResponse("de", settings)

	require.False(t, resp.AboutShow)
	require.Equal(t, "Über DE", resp.About)
}

func Test_MapSettingToSettingPageResponse_BooleanParsing(t *testing.T) {
	settings := []model.Setting{
		{Key: "about_show", LanguageCode: "en", Value: "TRUE"},
		{Key: "about", LanguageCode: "en", Value: "About"},
	}

	resp := MapSettingToSettingPageResponse("en", settings)

	// Only exact "true" is treated as true
	require.False(t, resp.AboutShow)
}

func Test_MapSettingToSettingPageResponse_MissingSettings(t *testing.T) {
	resp := MapSettingToSettingPageResponse("en", nil)

	require.NotNil(t, resp)

	require.False(t, resp.AboutShow)
	require.Empty(t, resp.About)

	require.False(t, resp.ContactShow)
	require.Empty(t, resp.Contact)

	require.False(t, resp.ImprintShow)
	require.Empty(t, resp.Imprint)

	require.False(t, resp.TermsOfUseShow)
	require.Empty(t, resp.TermsOfUse)

	require.False(t, resp.PrivacyPolicyShow)
	require.Empty(t, resp.PrivacyPolicy)
}
