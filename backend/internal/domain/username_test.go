package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/repository"
	"backend/internal/infrastructure/repository/mocks"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Unit_ValidateUsername(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	valid := []string{"abc", "a-b", "user-1", "a1b", "john-smith", strings.Repeat("a", 30), "x9y"}
	for _, name := range valid {
		require.NoError(t, validateUsername(name), name)
	}

	invalid := map[string]string{
		"":                      "empty",
		"ab":                    "too short",
		strings.Repeat("a", 31): "too long",
		"Abc":                   "uppercase",
		"-abc":                  "leading hyphen",
		"abc-":                  "trailing hyphen",
		"a_b":                   "underscore",
		"a b":                   "space",
		"a.b":                   "dot",
		"ünï":                   "non-ascii",
	}
	for name, why := range invalid {
		require.ErrorIs(t, validateUsername(name), ErrInvalidInput, "%q (%s)", name, why)
	}
}

func Test_Unit_ValidateUsername_RejectsEveryReservedWord(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	reserved := []string{
		// routes
		"app", "auth", "docs", "cloud", "about", "contact", "imprint", "privacy-policy", "terms-of-use",
		"api", "v1", "swagger", "health", "images",
		// generic
		"admin", "root", "support", "help", "static", "assets", "public", "login", "logout", "sign-in",
		"sign-up", "register", "settings", "dashboard", "profile", "user", "users", "me", "www", "null", "undefined",
	}
	for _, name := range reserved {
		err := validateUsername(name)
		require.ErrorIs(t, err, ErrInvalidInput, name)
		if len(name) >= usernameMinLength {
			require.ErrorContains(t, err, "reserved", name)
		}
	}
}

func Test_Unit_ValidateUsername_ReservedMatchesTheWholeNameOnly(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	require.NoError(t, validateUsername("api-docs"))
	require.NoError(t, validateUsername("my-admin"))
	require.NoError(t, validateUsername("admin1"))
}

func Test_Unit_ReservedNames_IncludeTheConfiguredAssetsBasePath(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("assets.basePath", "/uploads/files")

	require.True(t, isRouteReserved("uploads"))
	require.True(t, isRouteReserved("Uploads"))
	require.ErrorIs(t, validateUsername("uploads"), ErrInvalidInput)
	require.False(t, isRouteReserved("files"))
}

func Test_Unit_ReservedNames_RouteSubsetExcludesTheGenericOnes(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	// These block a username but are harmless as a top-level shelf path.
	for _, name := range []string{"admin", "support", "help", "root", "www", "me"} {
		require.True(t, isReservedUsername(name), name)
		require.False(t, isRouteReserved(name), name)
	}
	for _, name := range []string{"app", "auth", "docs", "cloud", "api", "v1", "images"} {
		require.True(t, isRouteReserved(name), name)
	}
	require.True(t, isRouteReserved("DOCS"))
}

func Test_Unit_SanitizeUsername(t *testing.T) {
	cases := map[string]string{
		"john.smith":     "john-smith",
		"John Smith":     "john-smith",
		"--a__b--":       "a-b",
		"a@b.c":          "a-b-c",
		"UPPER":          "upper",
		"already-fine":   "already-fine",
		"über":           "ber",
		"":               "",
		"!!!":            "",
		"a---b":          "a-b",
		"user.name+tag1": "user-name-tag1",
	}
	for raw, want := range cases {
		require.Equal(t, want, sanitizeUsername(raw), raw)
	}
}

func Test_Unit_EmailLocalPart(t *testing.T) {
	require.Equal(t, "myusername", emailLocalPart("myusername@domain.com"))
	require.Equal(t, "first.last", emailLocalPart("first.last@domain.com"))
	require.Equal(t, "no-at-sign", emailLocalPart("no-at-sign"))
	require.Equal(t, "a@b", emailLocalPart("a@b@c.com"))
	require.Equal(t, "", emailLocalPart("@domain.com"))
}

// usernameRepo returns a repository whose users already own the given names.
func usernameRepo(t *testing.T, taken ...string) *repository.Repository {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	set := map[string]bool{}
	for _, name := range taken {
		set[name] = true
	}
	userRepository := mocks.NewMockUserRepository(ctrl)
	userRepository.EXPECT().
		UsernameTaken(gomock.Any(), gomock.Any()).
		DoAndReturn(func(username, _ string) (bool, error) { return set[username], nil }).
		AnyTimes()
	return &repository.Repository{UserRepository: userRepository}
}

func Test_Unit_AvailableUsername(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	cases := []struct {
		name       string
		taken      []string
		candidates []string
		want       string
	}{
		{"uses the first candidate", nil, []string{"alice", "bob"}, "alice"},
		{"skips candidates that sanitize to nothing", nil, []string{"", "!!!", "bob"}, "bob"},
		{"falls back to member", nil, []string{"", "%%"}, "member"},
		{"falls back to member without candidates", nil, nil, "member"},
		{"sanitizes", nil, []string{"John.Smith"}, "john-smith"},
		{"appends -2 when taken", []string{"alice"}, []string{"alice"}, "alice-2"},
		{"keeps counting", []string{"alice", "alice-2", "alice-3"}, []string{"alice"}, "alice-4"},
		{"pads a name that is too short", nil, []string{"ab"}, "ab-2"},
		{"pads a single character", nil, []string{"a"}, "a-2"},
		{"moves off a reserved word", nil, []string{"admin"}, "admin-2"},
		{"moves off a reserved word that is also taken", []string{"admin-2"}, []string{"admin"}, "admin-3"},
		{"moves off a route", nil, []string{"docs"}, "docs-2"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := availableUsername(usernameRepo(t, c.taken...), c.candidates...)
			require.NoError(t, err)
			require.Equal(t, c.want, got)
			require.NoError(t, validateUsername(got))
		})
	}
}

func Test_Unit_AvailableUsername_TruncatesLongNamesAndKeepsTheSuffix(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	long := strings.Repeat("a", 40)

	got, err := availableUsername(usernameRepo(t), long)
	require.NoError(t, err)
	require.Equal(t, strings.Repeat("a", 30), got)

	got, err = availableUsername(usernameRepo(t, strings.Repeat("a", 30)), long)
	require.NoError(t, err)
	require.Equal(t, strings.Repeat("a", 28)+"-2", got)
	require.Len(t, got, 30)
}

func Test_Unit_AvailableUsername_TrimsAHyphenLeftByTruncation(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)

	// Cut at 30 characters this would end in a hyphen.
	raw := strings.Repeat("a", 29) + "-bbb"

	got, err := availableUsername(usernameRepo(t), raw)

	require.NoError(t, err)
	require.Equal(t, strings.Repeat("a", 29), got)
}

func Test_Unit_AvailableUsername_PropagatesRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	boom := errors.New("db unavailable")
	userRepository := mocks.NewMockUserRepository(ctrl)
	userRepository.EXPECT().UsernameTaken(gomock.Any(), gomock.Any()).Return(false, boom)

	_, err := availableUsername(&repository.Repository{UserRepository: userRepository}, "alice")

	require.ErrorIs(t, err, boom)
}

func Test_Unit_CheckUsernameAvailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	repo := &repository.Repository{UserRepository: userRepository}

	userRepository.EXPECT().UsernameTaken("alice", "user-1").Return(true, nil)
	require.ErrorIs(t, checkUsernameAvailable(repo, "alice", "user-1"), ErrConflict)

	userRepository.EXPECT().UsernameTaken("bob", "").Return(false, nil)
	require.NoError(t, checkUsernameAvailable(repo, "bob", ""))

	boom := errors.New("db unavailable")
	userRepository.EXPECT().UsernameTaken("carol", "").Return(false, boom)
	require.ErrorIs(t, checkUsernameAvailable(repo, "carol", ""), boom)
}
