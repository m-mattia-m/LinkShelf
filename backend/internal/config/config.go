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

var searchPaths = []string{".", "..", "../..", "../../..", "../../../..", "backend", "./backend"}

var k = koanf.New(".")

// LoadConfig loads configuration in three layers, each overriding the previous one:
//  1. config.default.yaml (or config.test.yaml when running under `go test`)
//  2. an optional config.yaml, if present
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

	if overridePath, err := findConfigFile("config.yaml"); err == nil {
		if err := k.Load(file.Provider(overridePath), yaml.Parser()); err != nil {
			return err
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
	if strings.TrimSpace(String("authentication.jwtSecret")) == "" {
		return fmt.Errorf("authentication.jwtSecret must be set")
	}

	switch strings.ToUpper(String("authentication.type")) {
	case "LOCAL":
	case "OIDC":
		if strings.TrimSpace(String("authentication.oidc.issuer")) == "" ||
			strings.TrimSpace(String("authentication.oidc.clientId")) == "" ||
			strings.TrimSpace(String("authentication.oidc.clientSecret")) == "" {
			return fmt.Errorf("authentication.oidc.issuer, clientId and clientSecret must be set when authentication.type is OIDC")
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

// Set overrides a single config value, mainly used by tests.
func Set(path string, val any) {
	_ = k.Set(path, val)
}

// Reset clears all loaded configuration, mainly used by tests.
func Reset() {
	k = koanf.New(".")
}
