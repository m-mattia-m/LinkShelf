package domain

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"backend/internal/infrastructure/repository/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func writeThemeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

func Test_Unit_SyncInstanceThemes_NoopWhenNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", "")

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_NoopWhenDirectoryDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", "/no/such/directory/linkshelf-test")

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_ValidFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()
	writeThemeFile(t, dir, "midnight.yaml", "name: Midnight\nconfig: |\n  --shelf-bg: #0f172a;\n")

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	themeRepository.EXPECT().
		UpsertInstanceBySourceFile(gomock.Any()).
		DoAndReturn(func(theme *model.Theme) error {
			require.Equal(t, "midnight", theme.SourceFile)
			require.Equal(t, "Midnight", theme.Name)
			require.Contains(t, theme.Config, "--shelf-bg: #0f172a;")
			return nil
		})

	themeRepository.EXPECT().
		DeleteInstanceNotIn([]string{"midnight"}).
		Return(nil)

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_SkipsInvalidYAML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()
	writeThemeFile(t, dir, "broken.yaml", "not: [valid: yaml")

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	// No UpsertInstanceBySourceFile call expected - the file is skipped.
	themeRepository.EXPECT().
		DeleteInstanceNotIn(gomock.Nil()).
		Return(nil)

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_SkipsMissingName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()
	writeThemeFile(t, dir, "noname.yaml", "config: |\n  --shelf-bg: #000;\n")

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	themeRepository.EXPECT().
		DeleteInstanceNotIn(gomock.Nil()).
		Return(nil)

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_SkipsInvalidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()
	writeThemeFile(t, dir, "bad-config.yaml", "name: Bad\nconfig: |\n  --shelf-bg: not-a-color;\n")

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	themeRepository.EXPECT().
		DeleteInstanceNotIn(gomock.Nil()).
		Return(nil)

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_SkipsDuplicateSlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()
	writeThemeFile(t, dir, "midnight.yaml", "name: Midnight\nconfig: |\n  --shelf-bg: #0f172a;\n")
	writeThemeFile(t, dir, "midnight.yml", "name: Midnight Two\nconfig: |\n  --shelf-bg: #000000;\n")

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	// Only one of the two same-slug files is synced - which one depends on
	// directory read order, but exactly one Upsert call must happen.
	themeRepository.EXPECT().
		UpsertInstanceBySourceFile(gomock.Any()).
		Return(nil).
		Times(1)

	themeRepository.EXPECT().
		DeleteInstanceNotIn([]string{"midnight"}).
		Return(nil)

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_IgnoresNonYamlFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()
	writeThemeFile(t, dir, "readme.txt", "not a theme file")

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	themeRepository.EXPECT().
		DeleteInstanceNotIn(gomock.Nil()).
		Return(nil)

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_ContinuesPastUpsertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()
	writeThemeFile(t, dir, "midnight.yaml", "name: Midnight\nconfig: |\n  --shelf-bg: #0f172a;\n")

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	themeRepository.EXPECT().
		UpsertInstanceBySourceFile(gomock.Any()).
		Return(errors.New("db unavailable"))

	// The failed file is not kept, so DeleteInstanceNotIn sees an empty list.
	themeRepository.EXPECT().
		DeleteInstanceNotIn(gomock.Nil()).
		Return(nil)

	require.NoError(t, SyncInstanceThemes(repo))
}

func Test_Unit_SyncInstanceThemes_PropagatesDeleteInstanceNotInError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	themeRepository.EXPECT().
		DeleteInstanceNotIn(gomock.Nil()).
		Return(errors.New("db unavailable"))

	err := SyncInstanceThemes(repo)

	require.ErrorContains(t, err, "db unavailable")
}

func Test_Unit_LoadInstanceThemeFile_Success(t *testing.T) {
	dir := t.TempDir()
	writeThemeFile(t, dir, "midnight.yaml", "name: Midnight\nconfig: |\n  --shelf-bg: #0f172a;\n")

	theme, err := loadInstanceThemeFile(filepath.Join(dir, "midnight.yaml"), "midnight")

	require.NoError(t, err)
	require.Equal(t, "Midnight", theme.Name)
	require.Equal(t, "midnight", theme.SourceFile)
	require.Equal(t, model.ThemeScopeInstance, theme.Scope)
	require.Contains(t, theme.Config, "--shelf-bg: #0f172a;")
}

func Test_Unit_LoadInstanceThemeFile_MissingName(t *testing.T) {
	dir := t.TempDir()
	writeThemeFile(t, dir, "noname.yaml", "config: |\n  --shelf-bg: #000;\n")

	_, err := loadInstanceThemeFile(filepath.Join(dir, "noname.yaml"), "noname")

	require.ErrorContains(t, err, "missing required")
}

func Test_Unit_LoadInstanceThemeFile_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	writeThemeFile(t, dir, "bad.yaml", "name: Bad\nconfig: |\n  --shelf-bg: not-a-color;\n")

	_, err := loadInstanceThemeFile(filepath.Join(dir, "bad.yaml"), "bad")

	require.ErrorContains(t, err, "invalid config")
}

func Test_Unit_LoadInstanceThemeFile_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	writeThemeFile(t, dir, "broken.yaml", "not: [valid: yaml")

	_, err := loadInstanceThemeFile(filepath.Join(dir, "broken.yaml"), "broken")

	require.ErrorContains(t, err, "parse yaml")
}

// bundledThemesDir is backend/themes - the "modern chic" theme pack
// config.default.yaml's themes.directory points at by default (see
// SyncInstanceThemes's doc comment).
func bundledThemesDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../themes")
	require.NoError(t, err)
	return dir
}

// Guards against a typo or an out-of-range value creeping into one of the
// bundled theme files: every file must load and validate the same way an
// admin-authored one would, and no two may claim the same display name.
func Test_Unit_BundledThemes_AllValidAndUniquelyNamed(t *testing.T) {
	dir := bundledThemesDir(t)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	names := make(map[string]bool)
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".yaml" {
			continue
		}
		sourceFile := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		theme, err := loadInstanceThemeFile(filepath.Join(dir, entry.Name()), sourceFile)

		require.NoError(t, err, "theme file %q", entry.Name())
		require.False(t, names[theme.Name], "duplicate theme name %q", theme.Name)
		names[theme.Name] = true
		count++
	}

	require.GreaterOrEqual(t, count, 5, "expected at least 5 bundled themes")
	// Not a real design constraint, just a sanity guard against a copy-paste
	// leaving duplicate-in-all-but-name files behind.
	require.LessOrEqual(t, count, 25, "expected at most 25 bundled themes")
}

// End-to-end through SyncInstanceThemes itself (not just the per-file
// loader), pointed at the real bundled directory rather than a temp one -
// proves config.default.yaml's own default actually works.
func Test_Unit_SyncInstanceThemes_BundledThemesDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := bundledThemesDir(t)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	themeRepository := mocks.NewMockThemeRepository(ctrl)
	repo := &repository.Repository{ThemeRepository: themeRepository}

	config.Reset()
	config.Set("themes.directory", dir)

	var upserted []string
	themeRepository.EXPECT().
		UpsertInstanceBySourceFile(gomock.Any()).
		DoAndReturn(func(theme *model.Theme) error {
			upserted = append(upserted, theme.SourceFile)
			return nil
		}).
		Times(len(entries))

	themeRepository.EXPECT().
		DeleteInstanceNotIn(gomock.Any()).
		Return(nil)

	require.NoError(t, SyncInstanceThemes(repo))
	require.Len(t, upserted, len(entries))
}
