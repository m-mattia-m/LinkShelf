package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadConfig_Success_ReadsDefaultsFromTestYaml(t *testing.T) {
	Reset()
	require.NoError(t, os.Unsetenv("APP_DATABASE_HOST"))

	require.NoError(t, LoadConfig())

	require.Equal(t, "LinkShelfTest", String("app.name"))
	require.Equal(t, "LOCAL", String("authentication.type"))
	require.NotEmpty(t, String("authentication.jwtSecret"))
	require.Equal(t, 5, Int("authentication.accessTokenExpiryMinutes"))
}

func Test_LoadConfig_EnvVarOverridesFile(t *testing.T) {
	Reset()
	require.NoError(t, os.Setenv("APP_DATABASE_HOST", "env-override-host"))
	defer func() { _ = os.Unsetenv("APP_DATABASE_HOST") }()

	require.NoError(t, LoadConfig())

	require.Equal(t, "env-override-host", String("database.host"))
}

func Test_LoadConfig_ConfigurationFilePathOverridesDefault(t *testing.T) {
	Reset()

	overridePath := filepath.Join(t.TempDir(), "config.override.yaml")
	require.NoError(t, os.WriteFile(overridePath, []byte("app:\n  name: OverriddenName\n"), 0644))

	require.NoError(t, os.Setenv("CONFIGURATION_FILE_PATH", overridePath))
	defer func() { _ = os.Unsetenv("CONFIGURATION_FILE_PATH") }()

	require.NoError(t, LoadConfig())

	require.Equal(t, "OverriddenName", String("app.name"))
	// Values the override file doesn't touch still come from the default.
	require.Equal(t, "LOCAL", String("authentication.type"))
}

func Test_LoadConfig_ConfigurationFilePathMissingFileFails(t *testing.T) {
	Reset()

	require.NoError(t, os.Setenv("CONFIGURATION_FILE_PATH", filepath.Join(t.TempDir(), "does-not-exist.yaml")))
	defer func() { _ = os.Unsetenv("CONFIGURATION_FILE_PATH") }()

	require.ErrorContains(t, LoadConfig(), "CONFIGURATION_FILE_PATH")
}

func Test_LoadConfig_NoConfigurationFilePathUsesDefaultOnly(t *testing.T) {
	Reset()
	require.NoError(t, os.Unsetenv("CONFIGURATION_FILE_PATH"))

	require.NoError(t, LoadConfig())

	require.Equal(t, "LinkShelfTest", String("app.name"))
}

func Test_Validate_FailsWithoutJwtSecret(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "")
	Set("authentication.type", "LOCAL")

	require.ErrorContains(t, validate(), "jwtSecret")
}

func Test_Validate_FailsForOidcWithoutIssuer(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "")
	Set("authentication.oidc.clientId", "")
	Set("authentication.oidc.clientSecret", "")

	require.ErrorContains(t, validate(), "OIDC")
}

func Test_Validate_SucceedsForOidcWithAllFields(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "https://issuer.example.com")
	Set("authentication.oidc.clientId", "client-id")
	Set("authentication.oidc.clientSecret", "client-secret")

	require.NoError(t, validate())
}

// Test_Validate_SucceedsForOidcWithoutClientSecret asserts that a public
// client using the PKCE flow - which needs no client secret - is a valid
// configuration; see oidcclient.Client.AuthorizationURL.
func Test_Validate_SucceedsForOidcWithoutClientSecret(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "https://issuer.example.com")
	Set("authentication.oidc.clientId", "client-id")
	Set("authentication.oidc.clientSecret", "")

	require.NoError(t, validate())
}

func Test_Validate_FailsForOidcWithoutClientId(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "https://issuer.example.com")
	Set("authentication.oidc.clientId", "")
	Set("authentication.oidc.clientSecret", "")

	require.ErrorContains(t, validate(), "OIDC")
}

func Test_Validate_FailsForUnknownAuthType(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "GOOGLE")

	require.ErrorContains(t, validate(), "unsupported authentication.type")
}

func Test_Validate_FailsForEmailVerificationEnabledWithoutSmtpHost(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", true)
	Set("smtp.host", "")
	Set("smtp.from", "no-reply@example.com")

	require.ErrorContains(t, validate(), "smtp.host")
}

func Test_Validate_FailsForEmailVerificationEnabledWithoutSmtpFrom(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", true)
	Set("smtp.host", "smtp.example.com")
	Set("smtp.from", "")

	require.ErrorContains(t, validate(), "smtp.from")
}

func Test_Validate_SucceedsForEmailVerificationEnabledWithSmtpConfigured(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", true)
	Set("smtp.host", "smtp.example.com")
	Set("smtp.from", "no-reply@example.com")

	require.NoError(t, validate())
}

func Test_Validate_SucceedsForEmailVerificationDisabledWithoutSmtp(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", false)

	require.NoError(t, validate())
}
