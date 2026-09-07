package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const envPrefix = "APP_"

// configFileEnvVar points to a second config file that overwrites the
// default, e.g. CONFIGURATION_FILE_PATH=config.prod.yaml. It is unprefixed
// (not APP_CONFIGURATION_FILE_PATH) since it selects which config to load
// rather than being a value within it, and is deliberately not picked up by
// the APP_-prefixed env.Provider below.
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
	Authentication struct {
		Type                      string `yaml:"type"`
		JwtSecret                 string `yaml:"jwtSecret" json:"-"`
		AccessTokenExpiryMinutes  int    `yaml:"accessTokenExpiryMinutes"`
		RefreshTokenExpiryMinutes int    `yaml:"refreshTokenExpiryMinutes"`
		BootstrapAdmin            struct {
			Email    string `yaml:"email"`
			Password string `yaml:"password" json:"-"`
		} `yaml:"bootstrapAdmin"`
		Oidc struct {
			Issuer       string `yaml:"issuer"`
			ClientId     string `yaml:"clientId"`
			ClientSecret string `yaml:"clientSecret" json:"-"`
			RedirectUrl  string `yaml:"redirectUrl"`
		} `yaml:"oidc"`
	} `yaml:"authentication"`
}

// LoadConfig loads configuration in three layers, each overriding the previous one:
//  1. config.default.yaml (or config.test.yaml when running under `go test`)
//  2. the file CONFIGURATION_FILE_PATH points to, if that env var is set
//  3. environment variables prefixed with APP_ (dots replace underscores, e.g. APP_DATABASE_HOST)
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

	if err := k.Load(env.Provider(envPrefix, ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, envPrefix)), "_", ".")
	}), nil); err != nil {
		return err
	}

	return validate()
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
