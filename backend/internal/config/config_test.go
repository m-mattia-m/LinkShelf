package config

import (
	"os"
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

func Test_Validate_FailsForUnknownAuthType(t *testing.T) {
	Reset()
	Set("authentication.jwtSecret", "some-secret")
	Set("authentication.type", "GOOGLE")

	require.ErrorContains(t, validate(), "unsupported authentication.type")
}
