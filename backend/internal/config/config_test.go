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

func validWithServiceToken(enabled bool, token string) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("authentication.serviceToken.enabled", enabled)
	Set("authentication.serviceToken.token", token)
}

func Test_Validate_ServiceTokenIsIgnoredWhenDisabled(t *testing.T) {
	validWithServiceToken(false, "")
	require.NoError(t, validate())
	require.False(t, ServiceTokenEnabled())
}

func Test_Validate_EnabledServiceTokenMustBeStrong(t *testing.T) {
	for _, token := range []string{
		"",
		"short",
		"0123456789012345678901234567890", // 31 bytes
		"change-me-to-a-long-random-value-in-production",
		testJwtSecret,
	} {
		validWithServiceToken(true, token)
		require.ErrorContains(t, validate(), "serviceToken.token", token)
	}
}

func Test_Validate_AcceptsAStrongServiceToken(t *testing.T) {
	validWithServiceToken(true, "01234567890123456789012345678901")
	require.NoError(t, validate())
	require.True(t, ServiceTokenEnabled())
}

func Test_LoadConfig_ServiceTokenIsOffByDefaultAndSettableFromTheEnvironment(t *testing.T) {
	Reset()
	require.NoError(t, LoadConfig())
	require.False(t, ServiceTokenEnabled())

	Reset()
	t.Setenv("AUTHENTICATION_SERVICETOKEN_ENABLED", "true")
	t.Setenv("AUTHENTICATION_SERVICETOKEN_TOKEN", "an-env-token-of-at-least-32-bytes-long")
	require.NoError(t, LoadConfig())
	require.True(t, ServiceTokenEnabled())
	require.Equal(t, "an-env-token-of-at-least-32-bytes-long", ServiceToken())
}

func Test_DefaultMaxShelves(t *testing.T) {
	five := 5
	zero := 0
	for name, tc := range map[string]struct {
		value   any
		want    *int
		wantErr bool
	}{
		"unset":    {value: nil},
		"empty":    {value: ""},
		"null":     {value: "null"},
		"int":      {value: 5, want: &five},
		"string":   {value: " 5 ", want: &five},
		"float":    {value: float64(5), want: &five},
		"zero":     {value: 0, want: &zero},
		"negative": {value: -1, wantErr: true},
		"text":     {value: "many", wantErr: true},
		"fraction": {value: 1.5, wantErr: true},
	} {
		t.Run(name, func(t *testing.T) {
			Reset()
			Set("limits.defaultMaxShelves", tc.value)
			got, err := DefaultMaxShelves()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_LoadConfig_DefaultsToUnlimitedShelves(t *testing.T) {
	Reset()
	require.NoError(t, LoadConfig())
	got, err := DefaultMaxShelves()
	require.NoError(t, err)
	require.Nil(t, got)

	Reset()
	t.Setenv("LIMITS_DEFAULTMAXSHELVES", "2")
	require.NoError(t, LoadConfig())
	got, err = DefaultMaxShelves()
	require.NoError(t, err)
	require.Equal(t, 2, *got)
}

func Test_Validate_RejectsANegativeDefaultMaxShelves(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("limits.defaultMaxShelves", -3)
	require.ErrorContains(t, validate(), "defaultMaxShelves")
}

func Test_Validate_UpgradeUrlMustBeHttp(t *testing.T) {
	for url, ok := range map[string]bool{
		"":                            true,
		"https://account.example.com": true,
		"javascript:alert(1)":         false,
		"ftp://example.com":           false,
		"/relative":                   false,
	} {
		Reset()
		Set("authentication.jwtSecret", testJwtSecret)
		Set("authentication.type", "LOCAL")
		Set("limits.upgradeUrl", url)
		if ok {
			require.NoError(t, validate(), url)
		} else {
			require.ErrorContains(t, validate(), "upgradeUrl", url)
		}
	}
}

func Test_SidebarLinks(t *testing.T) {
	Reset()
	links, err := SidebarLinks()
	require.NoError(t, err)
	require.Empty(t, links)

	Set("app.sidebarLinks", []map[string]any{{"name": " Status ", "url": "https://status.example.com", "icon": "i-lucide-activity"}})
	links, err = SidebarLinks()
	require.NoError(t, err)
	require.Equal(t, []SidebarLink{{Name: "Status", Url: "https://status.example.com", Icon: "i-lucide-activity"}}, links)

	for name, link := range map[string]map[string]any{
		"missing name": {"url": "https://example.com"},
		"js url":       {"name": "x", "url": "javascript:alert(1)"},
		"relative url": {"name": "x", "url": "/docs"},
	} {
		Set("app.sidebarLinks", []map[string]any{link})
		_, err = SidebarLinks()
		require.ErrorContains(t, err, "app.sidebarLinks", name)
	}
}

func oidcValidBase() {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "OIDC")
	Set("authentication.oidc.issuer", "https://issuer.example.com")
	Set("authentication.oidc.clientId", "client-id")
}

func Test_Validate_OidcAdditionalRedirectUrls(t *testing.T) {
	oidcValidBase()
	Set("authentication.oidc.additionalRedirectUrls", []string{"https://account.example.com/auth/callback", " http://localhost:3001/auth/callback "})
	require.NoError(t, validate())

	for name, bad := range map[string]string{
		"bad scheme": "ftp://account.example.com/cb",
		"no scheme":  "account.example.com/cb",
		"no host":    "https:///cb",
		"fragment":   "https://account.example.com/cb#frag",
		"empty":      "  ",
	} {
		t.Run(name, func(t *testing.T) {
			oidcValidBase()
			Set("authentication.oidc.additionalRedirectUrls", []string{bad})
			err := validate()
			require.Error(t, err)
			require.ErrorContains(t, err, "authentication.oidc.additionalRedirectUrls must be full http(s) URLs without a fragment")
		})
	}
}

func Test_OidcRedirectUrlAllowed(t *testing.T) {
	Reset()
	Set("authentication.oidc.redirectUrl", "https://app.example.com/auth/callback")
	Set("authentication.oidc.additionalRedirectUrls", []string{" https://account.example.com/auth/callback "})

	require.True(t, OidcRedirectUrlAllowed("https://app.example.com/auth/callback"))
	require.True(t, OidcRedirectUrlAllowed("https://account.example.com/auth/callback"))

	require.False(t, OidcRedirectUrlAllowed(""))
	require.False(t, OidcRedirectUrlAllowed("https://account.example.com/auth/callback/"))
	require.False(t, OidcRedirectUrlAllowed("https://account.example.com/auth/callback?x=1"))
	require.False(t, OidcRedirectUrlAllowed("https://ACCOUNT.example.com/auth/callback"))
	require.False(t, OidcRedirectUrlAllowed("https://account.example.com.evil.com/auth/callback"))
	require.False(t, OidcRedirectUrlAllowed("https://account.example.com"))
}

func Test_ServiceTokenScopes_DefaultToLimits(t *testing.T) {
	Reset()
	require.Equal(t, []string{ScopeLimits}, ServiceTokenScopes())
	require.True(t, ServiceTokenHasScope(ScopeLimits))
	require.False(t, ServiceTokenHasScope(ScopeDomainsRead))

	Set("authentication.serviceToken.scopes", []string{" ", ""})
	require.Equal(t, []string{ScopeLimits}, ServiceTokenScopes(), "blank entries count as none")
}

func Test_Validate_ServiceTokenScopesMustBeKnown(t *testing.T) {
	validWithServiceToken(true, "01234567890123456789012345678901")
	Set("authentication.serviceToken.scopes", []string{"limits", "domains:read", "domains:write-status"})
	require.NoError(t, validate())

	Set("authentication.serviceToken.scopes", []string{"limits", "domains:delete"})
	require.ErrorContains(t, validate(), `unknown scope "domains:delete"`)
}

func Test_LoadConfig_ServiceTokenScopesFromTheEnvironment(t *testing.T) {
	Reset()
	t.Setenv("AUTHENTICATION_SERVICETOKEN_ENABLED", "true")
	t.Setenv("AUTHENTICATION_SERVICETOKEN_TOKEN", "an-env-token-of-at-least-32-bytes-long")
	t.Setenv("AUTHENTICATION_SERVICETOKEN_SCOPES", "limits, domains:read")
	require.NoError(t, LoadConfig())
	require.Equal(t, []string{"limits", "domains:read"}, ServiceTokenScopes())
}

func Test_CustomDomainsRequireReady_OffByDefaultAndSettableFromTheEnvironment(t *testing.T) {
	Reset()
	require.NoError(t, LoadConfig())
	require.False(t, CustomDomainsRequireReady())

	Reset()
	t.Setenv("CUSTOMDOMAINS_REQUIREREADY", "true")
	require.NoError(t, LoadConfig())
	require.True(t, CustomDomainsRequireReady())
}

func validWithWebhooks(endpoints []map[string]any) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("webhooks.enabled", true)
	Set("webhooks.endpoints", endpoints)
}

func Test_WebhookEndpoints_OffUnlessEnabled(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", testJwtSecret)
	Set("authentication.type", "LOCAL")
	Set("webhooks.endpoints", []map[string]any{{"url": "https://example.com/hook", "secret": "x"}})
	require.Nil(t, WebhookEndpoints())
	require.NoError(t, validate())
}

func Test_WebhookEndpoints_SecretIsReadFromTheEnvironment(t *testing.T) {
	t.Setenv("WEBHOOKS_ACCOUNT_SECRET", " an-env-secret-of-at-least-32-bytes-long ")
	validWithWebhooks([]map[string]any{{"url": " https://account.example.com/v1/webhooks/linkshelf ", "secret": "${WEBHOOKS_ACCOUNT_SECRET}"}})

	require.NoError(t, validate())
	endpoints := WebhookEndpoints()
	require.Len(t, endpoints, 1)
	require.Equal(t, "https://account.example.com/v1/webhooks/linkshelf", endpoints[0].Url)
	require.Equal(t, "an-env-secret-of-at-least-32-bytes-long", endpoints[0].Secret)
	require.Equal(t, []string{"*"}, endpoints[0].Events, "no events means all of them")
}

func Test_Validate_WebhookEndpointsNeedAUrlAndAStrongSecret(t *testing.T) {
	strong := "an-env-secret-of-at-least-32-bytes-long"

	validWithWebhooks([]map[string]any{{"url": "ftp://example.com", "secret": strong}})
	require.ErrorContains(t, validate(), "webhooks.endpoints[0].url")

	validWithWebhooks([]map[string]any{{"url": "https://example.com/hook", "secret": "short"}})
	require.ErrorContains(t, validate(), "webhooks.endpoints[0].secret")

	// An unset variable leaves an empty secret, which is refused instead of signing with "".
	t.Setenv("WEBHOOKS_MISSING_SECRET", "")
	validWithWebhooks([]map[string]any{{"url": "https://example.com/hook", "secret": "${WEBHOOKS_MISSING_SECRET}"}})
	require.ErrorContains(t, validate(), "webhooks.endpoints[0].secret")

	validWithWebhooks([]map[string]any{{"url": "https://example.com/hook", "secret": strong, "events": []string{"shelf.domain.changed"}}})
	require.NoError(t, validate())
}
