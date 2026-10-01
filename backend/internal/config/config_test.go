package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/stretchr/testify/require"
)

// testJwtSecret passes validateJwtSecret: long enough and not a placeholder.
const testJwtSecret = "some-secret-that-is-at-least-32-bytes-long"

func Test_LoadConfig_Success_ReadsDefaultsFromTestYaml(t *testing.T) {
	Reset()
	require.NoError(t, os.Unsetenv("DATABASE_HOST"))

	require.NoError(t, LoadConfig())

	require.Equal(t, "LinkShelfTest", String("app.name"))
	require.Equal(t, "LOCAL", String("authentication.type"))
	require.NotEmpty(t, String("authentication.jwtSecret"))
	require.Equal(t, 5, Int("authentication.accessTokenExpiryMinutes"))
}

func Test_LoadConfig_EnvVarOverridesFile(t *testing.T) {
	Reset()
	require.NoError(t, os.Setenv("DATABASE_HOST", "env-override-host"))
	defer func() { _ = os.Unsetenv("DATABASE_HOST") }()

	require.NoError(t, LoadConfig())

	require.Equal(t, "env-override-host", String("database.host"))
}

func Test_LoadConfig_EnvVarOverridesCamelCaseKeys(t *testing.T) {
	Reset()
	t.Setenv("AUTHENTICATION_JWTSECRET", "from-env-a-random-value-of-at-least-32-bytes")
	t.Setenv("AUTHENTICATION_BOOTSTRAPADMIN_EMAIL", "env@example.com")
	t.Setenv("AUTHENTICATION_EMAILVERIFICATION_ENABLED", "false")
	t.Setenv("SMTP_TLSMODE", "starttls")

	require.NoError(t, LoadConfig())

	require.Equal(t, "from-env-a-random-value-of-at-least-32-bytes", String("authentication.jwtSecret"))
	require.Equal(t, "env@example.com", String("authentication.bootstrapAdmin.email"))
	require.False(t, Bool("authentication.emailVerification.enabled"))
	require.Equal(t, "starttls", String("smtp.tlsMode"))
	require.Equal(t, "from-env-a-random-value-of-at-least-32-bytes", Get().Authentication.JwtSecret)
}

func Test_LoadConfig_UserBasedPathsIsOffByDefaultAndSettableFromTheEnvironment(t *testing.T) {
	Reset()
	require.NoError(t, LoadConfig())
	require.False(t, Bool("app.userBasedPaths"))
	require.Equal(t, "admin", String("authentication.bootstrapAdmin.username"))

	Reset()
	t.Setenv("APP_USERBASEDPATHS", "true")
	t.Setenv("AUTHENTICATION_BOOTSTRAPADMIN_USERNAME", "root-admin")
	require.NoError(t, LoadConfig())

	require.True(t, Bool("app.userBasedPaths"))
	require.Equal(t, "root-admin", String("authentication.bootstrapAdmin.username"))
	require.True(t, Get().App.UserBasedPaths)
}

func Test_LoadConfig_EnvVarSplitsListValues(t *testing.T) {
	Reset()
	t.Setenv("SERVER_TRUSTEDPROXIES", "127.0.0.1, 10.0.0.0/8 ,")
	t.Setenv("APP_ADDITIONALORIGINS", "https://links.example.org,https://links.example.net")

	require.NoError(t, LoadConfig())

	require.Equal(t, []string{"127.0.0.1", "10.0.0.0/8"}, Strings("server.trustedProxies"))
	require.Equal(t, []string{"https://links.example.org", "https://links.example.net"}, Strings("app.additionalOrigins"))
}

func Test_LoadConfig_EnvVarSingleListValue(t *testing.T) {
	Reset()
	t.Setenv("APP_ADDITIONALORIGINS", "https://links.example.org")

	require.NoError(t, LoadConfig())

	require.Equal(t, []string{"https://links.example.org"}, Strings("app.additionalOrigins"))
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

// Placeholder and too-short secrets are refused.
func Test_Validate_RejectsWeakOrPlaceholderJwtSecrets(t *testing.T) {
	for _, secret := range []string{
		"change-me-to-a-long-random-value-in-production",
		"CHANGE-ME-TO-A-LONG-RANDOM-VALUE",
		"  change-me-to-a-long-random-value  ",
		"test-secret",
		"0123456789012345678901234567890", // 31 bytes
	} {
		Reset()
		Set("authentication.jwtSecret", secret)
		Set("authentication.type", "LOCAL")

		require.ErrorContains(t, validate(), "jwtSecret", secret)
	}
}

func Test_Validate_AcceptsA32ByteJwtSecret(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "01234567890123456789012345678901")
	Set("authentication.type", "LOCAL")

	require.NoError(t, validate())
}

func Test_LoadConfig_DefaultConfigShipsNoSecretAndNoBootstrapAdmin(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	path, err := findConfigFile("config.default.yaml")
	require.NoError(t, err)
	require.NoError(t, k.Load(file.Provider(path), yaml.Parser()))

	require.Empty(t, String("authentication.jwtSecret"))
	require.Empty(t, String("authentication.bootstrapAdmin.email"))
	require.Empty(t, String("authentication.bootstrapAdmin.password"))
	require.True(t, LocalAuthEnabled())
	require.ErrorContains(t, validate(), "jwtSecret")
}

func Test_Validate_LocalAuthCanOnlyBeDisabledWithOidc(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("authentication.localAuthEnabled", false)
	require.ErrorContains(t, validate(), "localAuthEnabled")

	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "https://issuer.example.com")
	Set("authentication.oidc.clientId", "client")
	Set("authentication.localAuthEnabled", false)
	require.NoError(t, validate())
}

func Test_LocalAuthEnabled_DefaultsToTrueWhenUnset(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	require.True(t, LocalAuthEnabled())

	Set("authentication.localAuthEnabled", false)
	require.False(t, LocalAuthEnabled())
}

func Test_Validate_FailsForOidcWithoutIssuer(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "")
	Set("authentication.oidc.clientId", "")
	Set("authentication.oidc.clientSecret", "")

	require.ErrorContains(t, validate(), "OIDC")
}

func Test_Validate_SucceedsForOidcWithAllFields(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "https://issuer.example.com")
	Set("authentication.oidc.clientId", "client-id")
	Set("authentication.oidc.clientSecret", "client-secret")

	require.NoError(t, validate())
}

// A public PKCE client needs no client secret.
func Test_Validate_SucceedsForOidcWithoutClientSecret(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "https://issuer.example.com")
	Set("authentication.oidc.clientId", "client-id")
	Set("authentication.oidc.clientSecret", "")

	require.NoError(t, validate())
}

func Test_Validate_FailsForOidcWithoutClientId(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "https://issuer.example.com")
	Set("authentication.oidc.clientId", "")
	Set("authentication.oidc.clientSecret", "")

	require.ErrorContains(t, validate(), "OIDC")
}

func Test_Validate_FailsForUnknownAuthType(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "GOOGLE")

	require.ErrorContains(t, validate(), "unsupported authentication.type")
}

func Test_Validate_FailsForEmailVerificationEnabledWithoutSmtpHost(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", true)
	Set("smtp.host", "")
	Set("smtp.from", "no-reply@example.com")

	require.ErrorContains(t, validate(), "smtp.host")
}

func Test_Validate_FailsForEmailVerificationEnabledWithoutSmtpFrom(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", true)
	Set("smtp.host", "smtp.example.com")
	Set("smtp.from", "")

	require.ErrorContains(t, validate(), "smtp.from")
}

func Test_Validate_SucceedsForEmailVerificationEnabledWithSmtpConfigured(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", true)
	Set("smtp.host", "smtp.example.com")
	Set("smtp.from", "no-reply@example.com")

	require.NoError(t, validate())
}

func Test_Validate_SucceedsForEmailVerificationDisabledWithoutSmtp(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", false)

	require.NoError(t, validate())
}

func Test_Validate_PasswordResetNeedsSmtp(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
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
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("authentication.passwordReset.enabled", false)
	Set("smtp.host", "")
	Set("smtp.from", "")

	require.NoError(t, validate())
}

func Test_Validate_StrictOriginsNeedsTheInstanceUrls(t *testing.T) {
	strict := func() {
		Reset()
		Set("authentication.jwtSecret", testJwtSecret)
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

func Test_Validate_StrictOriginsAdditionalOriginsMustBeFullUrls(t *testing.T) {
	strict := func() {
		Reset()
		Set("authentication.jwtSecret", testJwtSecret)
		Set("authentication.type", "LOCAL")
		Set("authentication.emailVerification.enabled", false)
		Set("authentication.passwordReset.enabled", false)
		Set("app.strictOrigins", true)
		Set("app.frontendUrl", "https://linkshelf.example.com")
		Set("server.host", "api.example.com")
	}

	strict()
	Set("app.additionalOrigins", []string{"https://links.example.org", "http://links.example.net"})
	require.NoError(t, validate())

	for _, bad := range []string{"linkshelf.example.com", "ftp://linkshelf.example.com", "https://", "://nope"} {
		strict()
		Set("app.additionalOrigins", []string{"https://links.example.org", bad})
		require.ErrorContains(t, validate(), "app.additionalOrigins", "additionalOrigins %q", bad)
	}
}

func Test_Validate_WithoutStrictOriginsTheUrlsMayBeEmpty(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("authentication.emailVerification.enabled", false)
	Set("authentication.passwordReset.enabled", false)
	Set("app.strictOrigins", false)
	Set("app.frontendUrl", "")
	Set("server.host", "")

	require.NoError(t, validate())
}
