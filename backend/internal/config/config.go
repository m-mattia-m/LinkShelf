package config

import (
	"flag"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// configFileEnvVar names an optional config file that overrides the default.
const configFileEnvVar = "CONFIGURATION_FILE_PATH"

var searchPaths = []string{".", "..", "../..", "../../..", "../../../..", "backend", "./backend"}

var k = koanf.New(".")

// Configuration mirrors config.default.yaml.
type Configuration struct {
	App struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Environment string `yaml:"environment"`
		Logo        string `yaml:"logo"`
		// FrontendUrl is the canonical address that email links point to.
		FrontendUrl string `yaml:"frontendUrl"`
		// AdditionalOrigins are other origins the frontend is served on.
		AdditionalOrigins []string `yaml:"additionalOrigins"`
		// UserBasedPaths switches shelf URLs from /<path> to /<username>/<path>.
		UserBasedPaths bool `yaml:"userBasedPaths"`
		// StrictOrigins restricts CORS to known origins and requests to Server.Host.
		StrictOrigins bool `yaml:"strictOrigins"`
	} `yaml:"app"`
	Server struct {
		Scheme         string   `yaml:"scheme"`
		Host           string   `yaml:"host"`
		Port           int      `yaml:"port"`
		TrustedProxies []string `yaml:"trustedProxies"`
	} `yaml:"server"`
	Database struct {
		Engine   string `yaml:"engine"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		// Password is excluded from JSON so it never ends up in a log line.
		Password string `yaml:"password" json:"-"`
		Name     string `yaml:"name"`
		Params   string `yaml:"params"`
	} `yaml:"database"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
	Domain struct {
		Openapi struct {
			UsePort bool `yaml:"usePort"`
		} `yaml:"openapi"`
	} `yaml:"domain"`
	Themes struct {
		// Directory of instance-theme YAML files, read at startup. Empty to skip.
		Directory string `yaml:"directory"`
	} `yaml:"themes"`
	Assets struct {
		// Directory of static files served at BasePath. Empty to skip.
		Directory string `yaml:"directory"`
		BasePath  string `yaml:"basePath"`
	} `yaml:"assets"`
	Authentication struct {
		Type                      string `yaml:"type"`
		JwtSecret                 string `yaml:"jwtSecret" json:"-"`
		AccessTokenExpiryMinutes  int    `yaml:"accessTokenExpiryMinutes"`
		RefreshTokenExpiryMinutes int    `yaml:"refreshTokenExpiryMinutes"`
		BootstrapAdmin            struct {
			Email    string `yaml:"email"`
			Password string `yaml:"password" json:"-"`
			Username string `yaml:"username"`
		} `yaml:"bootstrapAdmin"`
		Oidc struct {
			Issuer       string `yaml:"issuer"`
			ClientId     string `yaml:"clientId"`
			ClientSecret string `yaml:"clientSecret" json:"-"`
			RedirectUrl  string `yaml:"redirectUrl"`
		} `yaml:"oidc"`
		// LocalAuthEnabled gates password login, registration and reset. Only false with OIDC.
		LocalAuthEnabled bool `yaml:"localAuthEnabled"`
		// RegistrationEnabled gates self-registration only.
		RegistrationEnabled bool `yaml:"registrationEnabled"`
		EmailVerification   struct {
			Enabled bool `yaml:"enabled"`
			// TokenExpiryHours is how long a verification/invite link stays valid.
			TokenExpiryHours int `yaml:"tokenExpiryHours"`
		} `yaml:"emailVerification"`
		PasswordReset struct {
			Enabled bool `yaml:"enabled"`
			// TokenExpiryMinutes is how long an emailed reset link stays valid.
			TokenExpiryMinutes int `yaml:"tokenExpiryMinutes"`
		} `yaml:"passwordReset"`
		// ServiceToken is a static credential for the user-limits endpoints only.
		ServiceToken struct {
			Enabled bool   `yaml:"enabled"`
			Token   string `yaml:"token" json:"-"`
		} `yaml:"serviceToken"`
	} `yaml:"authentication"`
	Limits struct {
		// DefaultMaxShelves is copied onto new users. Empty means unlimited.
		DefaultMaxShelves any `yaml:"defaultMaxShelves"`
		// UpgradeUrl is shown when a user hits a limit. Optional.
		UpgradeUrl string `yaml:"upgradeUrl"`
	} `yaml:"limits"`
	Smtp struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password" json:"-"`
		From     string `yaml:"from"`
		// FromName is the optional display name in the From header.
		FromName string `yaml:"fromName"`
		// TlsMode is "none", "starttls" or "tls" (default).
		TlsMode string `yaml:"tlsMode"`
	} `yaml:"smtp"`
}

// LoadConfig loads config.default.yaml (config.test.yaml under go test), then
// CONFIGURATION_FILE_PATH, then env vars (database.host -> DATABASE_HOST).
func LoadConfig() error {
	baseName := "config.default.yaml"
	if isRunningTests() {
		baseName = "config.test.yaml"
	}

	basePath, err := findConfigFile(baseName)
	if err != nil {
		return err
	}
	if err := k.Load(file.Provider(basePath), yaml.Parser()); err != nil {
		return err
	}

	// An explicitly set path must exist.
	if path := strings.TrimSpace(os.Getenv(configFileEnvVar)); path != "" {
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return fmt.Errorf("load %s=%q: %w", configFileEnvVar, path, err)
		}
	}

	if err := k.Load(env.ProviderWithValue("", ".", envValueMapper(k)), nil); err != nil {
		return err
	}

	return validate()
}

// envKeyMapper maps an env var name to a config key, restoring the casing of known keys.
func envKeyMapper(existing []string) func(string) string {
	byLower := make(map[string]string, len(existing))
	for _, key := range existing {
		byLower[strings.ToLower(key)] = key
	}
	return func(s string) string {
		key := strings.ReplaceAll(strings.ToLower(s), "_", ".")
		if original, ok := byLower[key]; ok {
			return original
		}
		return key
	}
}

// envValueMapper splits list values on commas.
func envValueMapper(k *koanf.Koanf) func(key, value string) (string, interface{}) {
	toKey := envKeyMapper(k.Keys())
	return func(rawKey, rawValue string) (string, interface{}) {
		key := toKey(rawKey)
		switch k.Get(key).(type) {
		case []any, []string:
			parts := strings.Split(rawValue, ",")
			out := make([]string, 0, len(parts))
			for _, part := range parts {
				if part = strings.TrimSpace(part); part != "" {
					out = append(out, part)
				}
			}
			return key, out
		default:
			return key, rawValue
		}
	}
}

func validate() error {
	var cfg Configuration
	if err := k.Unmarshal("", &cfg); err != nil {
		return fmt.Errorf("config does not match the expected structure: %w", err)
	}

	if err := validateJwtSecret(String("authentication.jwtSecret")); err != nil {
		return err
	}
	if err := validateServiceToken(); err != nil {
		return err
	}
	if _, err := DefaultMaxShelves(); err != nil {
		return err
	}
	if err := validateUpgradeUrl(String("limits.upgradeUrl")); err != nil {
		return err
	}

	switch strings.ToUpper(String("authentication.type")) {
	case "LOCAL":
		if !LocalAuthEnabled() {
			return fmt.Errorf("authentication.localAuthEnabled can only be false when authentication.type is OIDC, otherwise nobody could sign in")
		}
	case "OIDC":
		// clientSecret is not required: the login flow always uses PKCE.
		if strings.TrimSpace(String("authentication.oidc.issuer")) == "" ||
			strings.TrimSpace(String("authentication.oidc.clientId")) == "" {
			return fmt.Errorf("authentication.oidc.issuer and clientId must be set when authentication.type is OIDC")
		}
	default:
		return fmt.Errorf("unsupported authentication.type %q, must be LOCAL or OIDC", String("authentication.type"))
	}

	if Bool("app.strictOrigins") {
		frontend, err := url.Parse(strings.TrimSpace(String("app.frontendUrl")))
		if err != nil || (frontend.Scheme != "http" && frontend.Scheme != "https") || frontend.Hostname() == "" {
			return fmt.Errorf("app.frontendUrl must be a full http(s) URL when app.strictOrigins is true, got %q", String("app.frontendUrl"))
		}
		for _, origin := range Strings("app.additionalOrigins") {
			origin = strings.TrimSpace(origin)
			parsed, err := url.Parse(origin)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
				return fmt.Errorf("app.additionalOrigins must be full http(s) URLs, got %q", origin)
			}
		}
		if strings.TrimSpace(String("server.host")) == "" {
			return fmt.Errorf("server.host must be set when app.strictOrigins is true")
		}
	}

	if Bool("authentication.emailVerification.enabled") {
		if strings.TrimSpace(String("smtp.host")) == "" || strings.TrimSpace(String("smtp.from")) == "" {
			return fmt.Errorf("smtp.host and smtp.from must be set when authentication.emailVerification.enabled is true")
		}
	}

	if Bool("authentication.passwordReset.enabled") {
		if strings.TrimSpace(String("smtp.host")) == "" || strings.TrimSpace(String("smtp.from")) == "" {
			return fmt.Errorf("smtp.host and smtp.from must be set when authentication.passwordReset.enabled is true")
		}
	}

	return nil
}

// minJwtSecretBytes matches the 256-bit HS256 key size.
const minJwtSecretBytes = 32

// knownPlaceholderJwtSecrets were shipped in configs or docs and are refused.
var knownPlaceholderJwtSecrets = []string{
	"change-me-to-a-long-random-value-in-production",
	"change-me-to-a-long-random-value",
}

func validateJwtSecret(secret string) error {
	return validateSecret("authentication.jwtSecret", secret)
}

func validateSecret(name, secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return fmt.Errorf("%s must be set, e.g. to the output of `openssl rand -base64 48`", name)
	}
	for _, placeholder := range knownPlaceholderJwtSecrets {
		if strings.EqualFold(secret, placeholder) {
			return fmt.Errorf("%s is a publicly known placeholder, set it to a random value, e.g. the output of `openssl rand -base64 48`", name)
		}
	}
	if len(secret) < minJwtSecretBytes {
		return fmt.Errorf("%s must be at least %d bytes long, e.g. the output of `openssl rand -base64 48`", name, minJwtSecretBytes)
	}
	return nil
}

func validateServiceToken() error {
	if !ServiceTokenEnabled() {
		return nil
	}
	token := ServiceToken()
	if err := validateSecret("authentication.serviceToken.token", token); err != nil {
		return err
	}
	if token == strings.TrimSpace(String("authentication.jwtSecret")) {
		return fmt.Errorf("authentication.serviceToken.token must differ from authentication.jwtSecret")
	}
	return nil
}

func validateUpgradeUrl(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return fmt.Errorf("limits.upgradeUrl must be a full http(s) URL, got %q", raw)
	}
	return nil
}

// ServiceTokenEnabled reports whether the service token is switched on (default: false).
func ServiceTokenEnabled() bool { return Bool("authentication.serviceToken.enabled") }

// ServiceToken returns the configured token, trimmed.
func ServiceToken() string { return strings.TrimSpace(String("authentication.serviceToken.token")) }

// DefaultMaxShelves returns the shelf limit for new users, nil for unlimited.
func DefaultMaxShelves() (*int, error) {
	const key = "limits.defaultMaxShelves"
	var n int
	switch v := k.Get(key).(type) {
	case nil:
		return nil, nil
	case int:
		n = v
	case int64:
		n = int(v)
	case float64:
		if v != math.Trunc(v) {
			return nil, fmt.Errorf("%s must be a whole number or empty, got %v", key, v)
		}
		n = int(v)
	case string:
		v = strings.TrimSpace(v)
		if v == "" || strings.EqualFold(v, "null") {
			return nil, nil
		}
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("%s must be a whole number or empty, got %q", key, v)
		}
		n = parsed
	default:
		return nil, fmt.Errorf("%s must be a whole number or empty, got %v", key, v)
	}
	if n < 0 {
		return nil, fmt.Errorf("%s must be zero or a positive number, got %d", key, n)
	}
	return &n, nil
}

// LocalAuthEnabled reports whether local auth is available (default: true).
func LocalAuthEnabled() bool {
	const key = "authentication.localAuthEnabled"
	return !k.Exists(key) || k.Bool(key)
}

func findConfigFile(name string) (string, error) {
	for _, dir := range searchPaths {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("config file %q not found", name)
}

func isRunningTests() bool {
	return flag.Lookup("test.v") != nil
}

// String, Bool, Int and Strings read a value at a dotted path.
func String(path string) string    { return k.String(path) }
func Bool(path string) bool        { return k.Bool(path) }
func Int(path string) int          { return k.Int(path) }
func Strings(path string) []string { return k.Strings(path) }

// Get returns the whole configuration; LoadConfig already validated it.
func Get() Configuration {
	var cfg Configuration
	_ = k.Unmarshal("", &cfg)
	return cfg
}

// Set overrides a single config value, mainly used by tests.
func Set(path string, val any) {
	_ = k.Set(path, val)
}

// Reset clears all loaded configuration, mainly used by tests.
func Reset() {
	k = koanf.New(".")
}
