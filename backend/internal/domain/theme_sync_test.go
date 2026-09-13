package domain

import (
	"errors"
	"os"
	"path/filepath"
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
