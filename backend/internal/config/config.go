package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// configFileEnvVar points to a second config file that overwrites the
// default, e.g. CONFIGURATION_FILE_PATH=config.prod.yaml. It selects which
// config to load rather than being a value within it, so it is read directly
// via os.Getenv instead of going through the env.Provider below.
const configFileEnvVar = "CONFIGURATION_FILE_PATH"

var searchPaths = []string{".", "..", "../..", "../../..", "../../../..", "backend", "./backend"}

var k = koanf.New(".")

// Configuration mirrors config.default.yaml so the whole tree can be read (or
// logged) as a typed value via Get(), instead of one dotted path at a time.
// mapstructure (which koanf's Unmarshal uses under the hood) matches these
// field names to yaml keys case-insensitively, so no struct tags are required
// for the mapping to work - they are kept here only for documentation.
type Configuration struct {
	App struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Environment string `yaml:"environment"`
		Logo        string `yaml:"logo"`
		// FrontendUrl is where verification/invite emails point their links
		// (e.g. "<FrontendUrl>/auth/verify-email?token=...") - the instance's
		// one canonical address.
		FrontendUrl string `yaml:"frontendUrl"`
		// AdditionalOrigins are other full http(s) origins the instance is
		// also reachable on (e.g. a second domain pointed at the same
		// frontend), besides FrontendUrl. Only used while StrictOrigins is
		// true: a browser calling the API from one of these is let through
		// the same way it would be from FrontendUrl's origin. A shelf's own
		// domain doesn't belong here, that is handled automatically.
		AdditionalOrigins []string `yaml:"additionalOrigins"`
		// UserBasedPaths switches public shelf URLs from /<path> to
		// /<username>/<path>.
		UserBasedPaths bool `yaml:"userBasedPaths"`
		// StrictOrigins locks the API to this instance's own hosts: browsers
		// may only call it from FrontendUrl's origin, one of
		// AdditionalOrigins, or a shelf's own domain, and it only answers
		// requests addressed to Server.Host.
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
		// Password is excluded from JSON so it never ends up in a log line
		// that dumps the whole configuration (e.g. main.go's startup log).
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
		// Directory of instance-theme YAML files, scanned once at startup
		// (see domain.SyncInstanceThemes). Leave empty to skip entirely.
		Directory string `yaml:"directory"`
	} `yaml:"themes"`
	Assets struct {
		// Directory of static files (e.g. theme background images) an
		// instance admin mounts in, scanned once at startup and served at
		// BasePath. Leave Directory empty to skip entirely.
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
		// RegistrationEnabled gates public self-registration (POST /v1/users
		// called without an admin token). Admin-created accounts and OIDC
		// auto-provisioning are never affected by this.
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
	} `yaml:"authentication"`
	Smtp struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password" json:"-"`
		From     string `yaml:"from"`
		// FromName is the optional display name shown alongside From, e.g.
		// "LinkShelf by Fermion" for a From of "no-reply@example.com" ->
		// `"LinkShelf by Fermion" <no-reply@example.com>` in the email
		// header. It is never sent as part of the SMTP envelope address
		// (MAIL FROM), only the header - some servers reject a display name
		// there. Left blank, the From header is just the bare address.
		FromName string `yaml:"fromName"`
		// TlsMode is "none", "starttls", or "tls" (implicit TLS). Anything
		// else (including blank/unset) is treated as "tls" - the secure
		// choice - by the mailer, not silently as "none".
		TlsMode string `yaml:"tlsMode"`
	} `yaml:"smtp"`
}

// LoadConfig loads configuration in three layers, each overriding the previous one:
//  1. config.default.yaml (or config.test.yaml when running under `go test`)
//  2. the file CONFIGURATION_FILE_PATH points to, if that env var is set
//  3. environment variables named after the key path, dots replaced by
//     underscores and upper-cased, e.g. database.host -> DATABASE_HOST
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

	// Overwrite the base with the file the env var points to. A path that was
	// set explicitly has to exist, otherwise a typo would silently start the
	// application with just the default configuration.
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

// envKeyMapper turns an environment variable name into a config key. Env var
// names cannot carry capitals, so AUTHENTICATION_JWTSECRET would otherwise
// become "authentication.jwtsecret" - a different koanf key than the
// "authentication.jwtSecret" every config.String() call reads. Matching the
// lowercased path against the keys the files already defined restores the
// original casing; unknown keys fall back to the plain lowercased path (which
// then does not match anything in the Configuration struct and is ignored).
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

// envValueMapper wraps envKeyMapper to also convert the value: a key that is
// a list in the already-loaded config (e.g. server.trustedProxies,
// app.additionalOrigins) is split on commas instead of becoming a single
// one-element string, which is what a plain env.Provider would otherwise do
// (silently discarding it, since a bare string isn't a []string). Blank
// entries, from a trailing comma or stray whitespace, are dropped.
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

	if strings.TrimSpace(String("authentication.jwtSecret")) == "" {
		return fmt.Errorf("authentication.jwtSecret must be set")
	}

	switch strings.ToUpper(String("authentication.type")) {
	case "LOCAL":
	case "OIDC":
		// clientSecret is intentionally not required: the login flow always
		// runs PKCE (see oidcclient.Client.AuthorizationURL), which is the
		// recommended flow for a public client and needs no client secret.
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

// String, Bool, Int and Strings read a config value at the given dotted path (e.g. "database.host").
func String(path string) string    { return k.String(path) }
func Bool(path string) bool        { return k.Bool(path) }
func Int(path string) int          { return k.Int(path) }
func Strings(path string) []string { return k.Strings(path) }

// Get returns the whole configuration as a typed value. The error is ignored
// because LoadConfig already unmarshals into Configuration once via
// validate(), so by the time Get() is called this cannot fail.
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
