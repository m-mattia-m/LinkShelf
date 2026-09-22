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

func Test_LoadConfig_EnvVarOverridesCamelCaseKeys(t *testing.T) {
	Reset()
	t.Setenv("APP_AUTHENTICATION_JWTSECRET", "from-env")
	t.Setenv("APP_AUTHENTICATION_BOOTSTRAPADMIN_EMAIL", "env@example.com")
	t.Setenv("APP_AUTHENTICATION_EMAILVERIFICATION_ENABLED", "false")
	t.Setenv("APP_SMTP_TLSMODE", "starttls")

	require.NoError(t, LoadConfig())

	require.Equal(t, "from-env", String("authentication.jwtSecret"))
	require.Equal(t, "env@example.com", String("authentication.bootstrapAdmin.email"))
	require.False(t, Bool("authentication.emailVerification.enabled"))
	require.Equal(t, "starttls", String("smtp.tlsMode"))
	require.Equal(t, "from-env", Get().Authentication.JwtSecret)
}

func Test_LoadConfig_UserBasedPathsIsOffByDefaultAndSettableFromTheEnvironment(t *testing.T) {
	Reset()
	require.NoError(t, LoadConfig())
	require.False(t, Bool("app.userBasedPaths"))
	require.Equal(t, "admin", String("authentication.bootstrapAdmin.username"))

	Reset()
	t.Setenv("APP_APP_USERBASEDPATHS", "true")
	t.Setenv("APP_AUTHENTICATION_BOOTSTRAPADMIN_USERNAME", "root-admin")
	require.NoError(t, LoadConfig())

	require.True(t, Bool("app.userBasedPaths"))
	require.Equal(t, "root-admin", String("authentication.bootstrapAdmin.username"))
	require.True(t, Get().App.UserBasedPaths)
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

func Test_Validate_PasswordResetNeedsSmtp(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "LOCAL")
	Set("authentication.passwordReset.enabled", true)
	Set("smtp.host", "")
	Set("smtp.from", "")

	require.ErrorContains(t, validate(), "authentication.passwordReset.enabled")

	Set("smtp.host", "smtp.example.com")
	require.ErrorContains(t, validate(), "smtp.from", "the sender is needed too")

	Set("smtp.from", "no-reply@example.com")
	require.NoError(t, validate())
}

func Test_Validate_PasswordResetDisabledNeedsNoSmtp(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "LOCAL")
	Set("authentication.passwordReset.enabled", false)
	Set("smtp.host", "")
	Set("smtp.from", "")

	require.NoError(t, validate())
}

func Test_Validate_StrictOriginsNeedsTheInstanceUrls(t *testing.T) {
	strict := func() {
		Reset()
		Set("authentication.jwtSecret", "some-secret")
		Set("authentication.type", "LOCAL")
		Set("authentication.emailVerification.enabled", false)
		Set("authentication.passwordReset.enabled", false)
		Set("app.strictOrigins", true)
		Set("app.frontendUrl", "https://linkshelf.example.com")
		Set("server.host", "api.example.com")
	}

	strict()
	require.NoError(t, validate())

	for _, bad := range []string{"", "   ", "linkshelf.example.com", "ftp://linkshelf.example.com", "https://", "://nope"} {
		strict()
		Set("app.frontendUrl", bad)
		require.ErrorContains(t, validate(), "app.frontendUrl", "frontendUrl %q", bad)
	}

	strict()
	Set("server.host", "")
	require.ErrorContains(t, validate(), "server.host")
}

func Test_Validate_WithoutStrictOriginsTheUrlsMayBeEmpty(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", false)
	Set("authentication.passwordReset.enabled", false)
	Set("app.strictOrigins", false)
	Set("app.frontendUrl", "")
	Set("server.host", "")

	require.NoError(t, validate())
}
