package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Unit_ValidateColor(t *testing.T) {
	valid := []string{"#fff", "#1c274c", "#1c274cff", "rgb(28, 39, 76)", "rgba(28, 39, 76, 0.5)", "transparent", "currentColor", "black", "white"}
	for _, value := range valid {
		t.Run(value, func(t *testing.T) {
			require.NoError(t, validateColor(value))
		})
	}

	invalid := []string{"", "not-a-color", "red", "url(javascript:alert(1))", "#12", strings.Repeat("a", 400)}
	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			require.ErrorIs(t, validateColor(value), ErrInvalidInput)
		})
	}
}

func Test_Unit_ValidateColorOrGradient(t *testing.T) {
	require.NoError(t, validateColorOrGradient("linear-gradient(180deg, #1c274c, #ffffff)"))
	require.NoError(t, validateColorOrGradient("radial-gradient(circle, #000, #fff)"))
	require.NoError(t, validateColorOrGradient("#1c274c"))
	require.ErrorIs(t, validateColorOrGradient("not-a-color-or-gradient"), ErrInvalidInput)
	require.ErrorIs(t, validateColorOrGradient("url(javascript:alert(1))"), ErrInvalidInput)
}

func Test_Unit_ValidateLength(t *testing.T) {
	valid := []string{"12px", "0.5rem", "50%", "9999px", "1em"}
	for _, value := range valid {
		t.Run(value, func(t *testing.T) {
			require.NoError(t, validateLength(value))
		})
	}

	invalid := []string{"", "12", "px", "12pixels", "-5px", "12px;color:red"}
	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			require.ErrorIs(t, validateLength(value), ErrInvalidInput)
		})
	}
}

func Test_Unit_ValidateFontFamily(t *testing.T) {
	require.NoError(t, validateFontFamily("'Inter', sans-serif"))
	require.NoError(t, validateFontFamily("Arial"))

	require.ErrorIs(t, validateFontFamily(""), ErrInvalidInput)
	require.ErrorIs(t, validateFontFamily("Inter; } body { display: none"), ErrInvalidInput)
	require.ErrorIs(t, validateFontFamily(strings.Repeat("a", 250)), ErrInvalidInput)
}

func Test_Unit_ValidateImageReference(t *testing.T) {
	require.NoError(t, validateImageReference("https://example.com/bg.png"))
	require.NoError(t, validateImageReference("http://example.com/bg.png"))
	require.NoError(t, validateImageReference("/images/my-dog.webp"))

	invalid := []string{
		"",
		"javascript:alert(1)",
		"data:text/html,<script>alert(1)</script>",
		"file:///etc/passwd",
		"images/relative.png",
		strings.Repeat("a", 400),
	}
	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			require.ErrorIs(t, validateImageReference(value), ErrInvalidInput)
		})
	}
}

func Test_Unit_ParseThemeConfig_Success(t *testing.T) {
	values, err := ParseThemeConfig(`
		--shelf-bg: #1c274c;
		--shelf-text: #ffffff;
		--shelf-link-radius: 12px;
	`)

	require.NoError(t, err)
	require.Equal(t, "#1c274c", values["--shelf-bg"])
	require.Equal(t, "#ffffff", values["--shelf-text"])
	require.Equal(t, "12px", values["--shelf-link-radius"])
	require.Len(t, values, 3)
}

func Test_Unit_ParseThemeConfig_IgnoresBlankDeclarations(t *testing.T) {
	values, err := ParseThemeConfig("--shelf-bg: #1c274c;;;  ;")

	require.NoError(t, err)
	require.Len(t, values, 1)
}

func Test_Unit_ParseThemeConfig_UnknownProperty(t *testing.T) {
	_, err := ParseThemeConfig("--not-a-real-property: red;")

	require.ErrorIs(t, err, ErrInvalidInput)
	require.ErrorContains(t, err, "unknown theme property")
}

func Test_Unit_ParseThemeConfig_DuplicateProperty(t *testing.T) {
	_, err := ParseThemeConfig("--shelf-bg: #000; --shelf-bg: #fff;")

	require.ErrorIs(t, err, ErrInvalidInput)
	require.ErrorContains(t, err, "more than once")
}

func Test_Unit_ParseThemeConfig_MalformedDeclaration(t *testing.T) {
	_, err := ParseThemeConfig("this is not a declaration")

	require.ErrorIs(t, err, ErrInvalidInput)
}

func Test_Unit_ParseThemeConfig_InvalidValue(t *testing.T) {
	_, err := ParseThemeConfig("--shelf-bg: not-a-color;")

	require.ErrorIs(t, err, ErrInvalidInput)
	require.ErrorContains(t, err, "--shelf-bg")
}

func Test_Unit_ParseThemeConfig_EmptyConfig(t *testing.T) {
	_, err := ParseThemeConfig("   ")

	require.ErrorIs(t, err, ErrInvalidInput)
	require.ErrorContains(t, err, "no valid properties")
}

func Test_Unit_SerializeThemeConfig_OrderAndFiltering(t *testing.T) {
	serialized := SerializeThemeConfig(map[string]string{
		"--shelf-font-family": "'Inter', sans-serif",
		"--shelf-bg":          "#1c274c",
		"--not-a-property":    "ignored",
	})

	bgIndex := strings.Index(serialized, "--shelf-bg")
	fontIndex := strings.Index(serialized, "--shelf-font-family")

	require.GreaterOrEqual(t, bgIndex, 0)
	require.GreaterOrEqual(t, fontIndex, 0)
	require.Less(t, bgIndex, fontIndex, "serialization should follow ThemeProperties' declared order")
	require.NotContains(t, serialized, "--not-a-property")
}

func Test_Unit_ValidateThemeConfig_Success(t *testing.T) {
	canonical, err := ValidateThemeConfig("--shelf-bg: #1c274c;")

	require.NoError(t, err)
	require.Contains(t, canonical, "--shelf-bg: #1c274c;")
}

func Test_Unit_ValidateThemeConfig_PropagatesParseError(t *testing.T) {
	_, err := ValidateThemeConfig("--shelf-bg: not-a-color;")

	require.ErrorIs(t, err, ErrInvalidInput)
}
