package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// ThemeProperties is the fixed set of CSS custom properties a theme may set.
// This is an allowlist, not a denylist: any property not listed here is
// rejected outright, and no raw CSS (selectors, at-rules, arbitrary
// declarations) is ever accepted. That's what makes user-authored themes safe
// to render on public pages without sandboxing or a CSS parser - the values
// are constrained by the validators below, and callers re-serialize the
// parsed result themselves rather than ever storing/replaying the caller's
// raw text verbatim.
var ThemeProperties = []string{
	"--shelf-bg",
	"--shelf-text",
	"--shelf-link-bg",
	"--shelf-link-text",
	"--shelf-link-radius",
	"--shelf-font-family",
	"--shelf-bg-image",
}

const maxThemeValueLength = 300

var themePropertyValidators = map[string]func(value string) error{
	"--shelf-bg":          validateColorOrGradient,
	"--shelf-text":        validateColor,
	"--shelf-link-bg":     validateColor,
	"--shelf-link-text":   validateColor,
	"--shelf-link-radius": validateLength,
	"--shelf-font-family": validateFontFamily,
	"--shelf-bg-image":    validateImageReference,
}

var (
	hexColorPattern   = regexp.MustCompile(`^#[0-9a-fA-F]{3,8}$`)
	colorFnPattern    = regexp.MustCompile(`^(rgb|rgba|hsl|hsla)\([0-9.%,\s]+\)$`)
	gradientFnPattern = regexp.MustCompile(`^(linear|radial)-gradient\([0-9a-zA-Z#.,%\s-]+\)$`)
	colorKeywords     = map[string]bool{
		"transparent": true, "currentcolor": true, "inherit": true, "black": true, "white": true,
	}
	lengthPattern     = regexp.MustCompile(`^\d+(\.\d+)?(px|rem|em|%)$`)
	fontFamilyPattern = regexp.MustCompile(`^[a-zA-Z0-9\s,'"-]+$`)
	imagePathPattern  = regexp.MustCompile(`^(https?://[^\s'"<>]+|/images/[a-zA-Z0-9._-]+)$`)
)

// hasNoInjectionCharacters rejects characters that would let a value break
// out of a single "property: value;" declaration when it's later
// re-serialized, or that have no legitimate use in any of the values below
// (url(...) is only ever allowed via --shelf-bg-image's own validator, never
// as a generic escape hatch in a color/length/font value).
func hasNoInjectionCharacters(value string) bool {
	if len(value) == 0 || len(value) > maxThemeValueLength {
		return false
	}
	return !strings.ContainsAny(value, ";{}<>\\") && !strings.Contains(strings.ToLower(value), "url(")
}

func validateColor(value string) error {
	if !hasNoInjectionCharacters(value) {
		return fmt.Errorf("%w: not a valid color", ErrInvalidInput)
	}
	if hexColorPattern.MatchString(value) || colorFnPattern.MatchString(value) || colorKeywords[strings.ToLower(value)] {
		return nil
	}
	return fmt.Errorf("%w: %q is not a valid color", ErrInvalidInput, value)
}

func validateColorOrGradient(value string) error {
	if !hasNoInjectionCharacters(value) {
		return fmt.Errorf("%w: not a valid color or gradient", ErrInvalidInput)
	}
	if gradientFnPattern.MatchString(value) {
		return nil
	}
	return validateColor(value)
}

func validateLength(value string) error {
	if !hasNoInjectionCharacters(value) || !lengthPattern.MatchString(value) {
		return fmt.Errorf("%w: %q is not a valid length (expected e.g. 12px, 0.5rem, 50%%)", ErrInvalidInput, value)
	}
	return nil
}

func validateFontFamily(value string) error {
	if !hasNoInjectionCharacters(value) || len(value) > 200 || !fontFamilyPattern.MatchString(value) {
		return fmt.Errorf("%w: %q is not a valid font family", ErrInvalidInput, value)
	}
	return nil
}

// validateImageReference is the one property allowed to carry a URL - either
// an absolute http(s) URL or a same-origin path served from the instance's
// mounted assets directory (see config "assets.directory"). Anything else
// (javascript:, data:, file:, or a bare relative path) is rejected.
func validateImageReference(value string) error {
	if len(value) == 0 || len(value) > maxThemeValueLength {
		return fmt.Errorf("%w: not a valid image reference", ErrInvalidInput)
	}
	if !imagePathPattern.MatchString(value) {
		return fmt.Errorf("%w: %q must be an http(s) URL or an /images/... path", ErrInvalidInput, value)
	}
	return nil
}

// ParseThemeConfig parses a theme's raw "--name: value;" declaration text
// into a validated map, rejecting any property not in ThemeProperties and any
// value that fails its property-specific validator. Duplicate properties are
// rejected rather than silently taking the last one, so authors get clear
// feedback instead of surprising behavior.
func ParseThemeConfig(raw string) (map[string]string, error) {
	values := make(map[string]string)

	for _, decl := range strings.Split(raw, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}

		parts := strings.SplitN(decl, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: %q is not a valid \"--property: value\" declaration", ErrInvalidInput, decl)
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		validator, ok := themePropertyValidators[name]
		if !ok {
			return nil, fmt.Errorf("%w: unknown theme property %q", ErrInvalidInput, name)
		}
		if _, exists := values[name]; exists {
			return nil, fmt.Errorf("%w: theme property %q is set more than once", ErrInvalidInput, name)
		}
		if err := validator(value); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		values[name] = value
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("%w: theme config has no valid properties", ErrInvalidInput)
	}

	return values, nil
}

// SerializeThemeConfig re-emits a validated property map as canonical
// "--property: value;\n" text, in a fixed order, so what's stored/exported is
// always exactly what was validated - never the caller's original formatting.
func SerializeThemeConfig(values map[string]string) string {
	var b strings.Builder
	for _, name := range ThemeProperties {
		if value, ok := values[name]; ok {
			b.WriteString(fmt.Sprintf("%s: %s;\n", name, value))
		}
	}
	return b.String()
}

// ValidateThemeConfig parses and re-serializes raw theme config text,
// returning the canonical form to store, or an error naming the first
// problem found.
func ValidateThemeConfig(raw string) (string, error) {
	values, err := ParseThemeConfig(raw)
	if err != nil {
		return "", err
	}
	return SerializeThemeConfig(values), nil
}
