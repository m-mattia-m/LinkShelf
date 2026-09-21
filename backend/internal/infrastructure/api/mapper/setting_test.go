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

	resp := MapSettingToSettingPageResponse("en", settings, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: false})

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

	resp := MapSettingToSettingPageResponse("de", settings, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: false})

	require.False(t, resp.Body.AboutShow)
	require.Equal(t, "Über DE", resp.Body.About)
}

func Test_MapSettingToSettingPageResponse_BooleanParsing(t *testing.T) {
	settings := []model.Setting{
		{Key: "about_show", LanguageCode: "en", Value: "TRUE"},
		{Key: "about", LanguageCode: "en", Value: "About"},
	}

	resp := MapSettingToSettingPageResponse("en", settings, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: false})

	// Only exact "true" is treated as true
	require.False(t, resp.Body.AboutShow)
}

func Test_MapSettingToSettingPageResponse_MissingSettings(t *testing.T) {
	resp := MapSettingToSettingPageResponse("en", nil, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: false})

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

func Test_MapSettingToSettingPageResponse_OidcEnabled(t *testing.T) {
	require.True(t, MapSettingToSettingPageResponse("en", nil, SettingPageFlags{OidcEnabled: true, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: false}).Body.OidcEnabled)
	require.False(t, MapSettingToSettingPageResponse("en", nil, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: false}).Body.OidcEnabled)
}

func Test_MapSettingToSettingPageResponse_RegistrationAndEmailVerificationEnabled(t *testing.T) {
	resp := MapSettingToSettingPageResponse("en", nil, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: true, EmailVerificationEnabled: true, UserBasedPaths: false})

	require.True(t, resp.Body.RegistrationEnabled)
	require.True(t, resp.Body.EmailVerificationEnabled)

	resp = MapSettingToSettingPageResponse("en", nil, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: false})

	require.False(t, resp.Body.RegistrationEnabled)
	require.False(t, resp.Body.EmailVerificationEnabled)
}

func Test_MapSettingToSettingPageResponse_UserBasedPaths(t *testing.T) {
	require.True(t, MapSettingToSettingPageResponse("en", nil, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: true}).Body.UserBasedPaths)
	require.False(t, MapSettingToSettingPageResponse("en", nil, SettingPageFlags{OidcEnabled: false, RegistrationEnabled: false, EmailVerificationEnabled: false, UserBasedPaths: false}).Body.UserBasedPaths)
}

func Test_MapSettingToSettingPageResponse_PasswordResetEnabled(t *testing.T) {
	require.True(t, MapSettingToSettingPageResponse("en", nil, SettingPageFlags{PasswordResetEnabled: true}).Body.PasswordResetEnabled)
	require.False(t, MapSettingToSettingPageResponse("en", nil, SettingPageFlags{}).Body.PasswordResetEnabled)
}
