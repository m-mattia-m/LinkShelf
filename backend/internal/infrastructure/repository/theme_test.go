package repository

import (
	"backend/internal/infrastructure/api/model"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_ThemeRepository_Get_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id", "scope", "owner_user_id", "name", "source_file", "config"}).
		AddRow("theme-uuid-test", "user", "user-uuid-test", "Sunset", nil, "--shelf-bg: #1c274c;\n")

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("theme-uuid-test").
		WillReturnRows(rows)

	theme, err := repo.Get("theme-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, theme)
	require.Equal(t, "Sunset", theme.Name)
	require.Equal(t, "user-uuid-test", theme.OwnerUserId)
	require.Empty(t, theme.SourceFile)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_Get_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("missing-theme").
		WillReturnError(sql.ErrNoRows)

	theme, err := repo.Get("missing-theme")

	require.NoError(t, err)
	require.Nil(t, theme)
}

func Test_ThemeRepository_Get_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id"}).AddRow("only-id")

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("theme-uuid-test").
		WillReturnRows(rows)

	theme, err := repo.Get("theme-uuid-test")

	require.Error(t, err)
	require.Nil(t, theme)
}

func Test_ThemeRepository_GetBySourceFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id", "scope", "owner_user_id", "name", "source_file", "config"}).
		AddRow("theme-uuid-test", "instance", nil, "Midnight", "midnight", "--shelf-bg: #0f172a;\n")

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("midnight").
		WillReturnRows(rows)

	theme, err := repo.GetBySourceFile("midnight")

	require.NoError(t, err)
	require.NotNil(t, theme)
	require.Equal(t, "Midnight", theme.Name)
	require.Equal(t, "midnight", theme.SourceFile)
	require.Empty(t, theme.OwnerUserId)
}

func Test_ThemeRepository_GetBySourceFile_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	theme, err := repo.GetBySourceFile("missing")

	require.NoError(t, err)
	require.Nil(t, theme)
}

func Test_ThemeRepository_ListInstance_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id", "scope", "owner_user_id", "name", "source_file", "config"}).
		AddRow("theme-1", "instance", nil, "Midnight", "midnight", "--shelf-bg: #000;\n")

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("instance").
		WillReturnRows(rows)

	themes, err := repo.ListInstance()

	require.NoError(t, err)
	require.Len(t, themes, 1)
	require.Equal(t, "Midnight", themes[0].Name)
}

func Test_ThemeRepository_ListInstance_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("instance").
		WillReturnError(errors.New("db unavailable"))

	themes, err := repo.ListInstance()

	require.Error(t, err)
	require.Nil(t, themes)
}

func Test_ThemeRepository_ListByOwner_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id", "scope", "owner_user_id", "name", "source_file", "config"}).
		AddRow("theme-1", "user", "user-uuid-test", "Sunset", nil, "--shelf-bg: #000;\n")

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("user", "user-uuid-test").
		WillReturnRows(rows)

	themes, err := repo.ListByOwner("user-uuid-test")

	require.NoError(t, err)
	require.Len(t, themes, 1)
	require.Equal(t, "user-uuid-test", themes[0].OwnerUserId)
}

func Test_ThemeRepository_ListAllUserScoped_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id", "scope", "owner_user_id", "name", "source_file", "config"}).
		AddRow("theme-1", "user", "user-1", "A", nil, "--shelf-bg: #000;\n").
		AddRow("theme-2", "user", "user-2", "B", nil, "--shelf-bg: #fff;\n")

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("user").
		WillReturnRows(rows)

	themes, err := repo.ListAllUserScoped()

	require.NoError(t, err)
	require.Len(t, themes, 2)
}

func Test_ThemeRepository_ListAllUserScoped_RowsIterationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id", "scope", "owner_user_id", "name", "source_file", "config"}).
		AddRow("theme-1", "user", "user-1", "A", nil, "--shelf-bg: #000;\n").
		RowError(0, errors.New("connection dropped mid-stream"))

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("user").
		WillReturnRows(rows)

	themes, err := repo.ListAllUserScoped()

	require.Error(t, err)
	require.Nil(t, themes)
}

func Test_ThemeRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("INSERT INTO theme").
		WithArgs(
			sqlmock.AnyArg(), // generated UUID
			model.ThemeScopeUser,
			"user-uuid-test",
			"Sunset",
			nil,
			"--shelf-bg: #1c274c;\n",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	theme := &model.Theme{
		Scope:       model.ThemeScopeUser,
		OwnerUserId: "user-uuid-test",
		Name:        "Sunset",
		Config:      "--shelf-bg: #1c274c;\n",
	}

	id, err := repo.Create(theme)

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.Equal(t, id, theme.Id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_Create_NullsEmptyOwnerAndSourceFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("INSERT INTO theme").
		WithArgs(
			sqlmock.AnyArg(),
			model.ThemeScopeInstance,
			nil,
			"Midnight",
			"midnight",
			"--shelf-bg: #0f172a;\n",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	theme := &model.Theme{
		Scope:      model.ThemeScopeInstance,
		Name:       "Midnight",
		SourceFile: "midnight",
		Config:     "--shelf-bg: #0f172a;\n",
	}

	_, err = repo.Create(theme)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_Create_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("INSERT INTO theme").
		WillReturnError(errors.New("insert failed"))

	id, err := repo.Create(&model.Theme{})

	require.Error(t, err)
	require.Empty(t, id)
}

func Test_ThemeRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("UPDATE theme").
		WithArgs("Updated Name", "--shelf-bg: #fff;\n", "theme-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(&model.Theme{Id: "theme-uuid-test", Name: "Updated Name", Config: "--shelf-bg: #fff;\n"})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_Update_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("UPDATE theme").
		WillReturnError(errors.New("update failed"))

	err = repo.Update(&model.Theme{Id: "theme-uuid-test"})

	require.Error(t, err)
}

func Test_ThemeRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("DELETE FROM theme").
		WithArgs("theme-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete("theme-uuid-test")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_Delete_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("DELETE FROM theme").
		WillReturnError(errors.New("delete failed"))

	err = repo.Delete("theme-uuid-test")

	require.Error(t, err)
}

func Test_ThemeRepository_UpsertInstanceBySourceFile_Creates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("midnight").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec("INSERT INTO theme").
		WithArgs(sqlmock.AnyArg(), model.ThemeScopeInstance, nil, "Midnight", "midnight", "--shelf-bg: #0f172a;\n").
		WillReturnResult(sqlmock.NewResult(1, 1))

	theme := &model.Theme{Name: "Midnight", SourceFile: "midnight", Config: "--shelf-bg: #0f172a;\n"}

	err = repo.UpsertInstanceBySourceFile(theme)

	require.NoError(t, err)
	require.Equal(t, model.ThemeScopeInstance, theme.Scope)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_UpsertInstanceBySourceFile_Updates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	existingRows := sqlmock.NewRows([]string{"id", "scope", "owner_user_id", "name", "source_file", "config"}).
		AddRow("existing-theme-uuid", "instance", nil, "Midnight (old)", "midnight", "--shelf-bg: #000;\n")

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("midnight").
		WillReturnRows(existingRows)

	mock.ExpectExec("UPDATE theme").
		WithArgs("Midnight", "--shelf-bg: #0f172a;\n", "existing-theme-uuid").
		WillReturnResult(sqlmock.NewResult(1, 1))

	theme := &model.Theme{Name: "Midnight", SourceFile: "midnight", Config: "--shelf-bg: #0f172a;\n"}

	err = repo.UpsertInstanceBySourceFile(theme)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_UpsertInstanceBySourceFile_GetError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+theme`).
		WithArgs("midnight").
		WillReturnError(errors.New("db unavailable"))

	err = repo.UpsertInstanceBySourceFile(&model.Theme{SourceFile: "midnight"})

	require.Error(t, err)
}

func Test_ThemeRepository_DeleteInstanceNotIn_EmptyDeletesAllInstance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("DELETE FROM theme").
		WithArgs(model.ThemeScopeInstance).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err = repo.DeleteInstanceNotIn(nil)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_DeleteInstanceNotIn_KeepsListed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("DELETE FROM theme").
		WithArgs(model.ThemeScopeInstance, "midnight", "sunset").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteInstanceNotIn([]string{"midnight", "sunset"})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ThemeRepository_DeleteInstanceNotIn_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &themeRepository{Engine: db}

	mock.ExpectExec("DELETE FROM theme").
		WillReturnError(errors.New("delete failed"))

	err = repo.DeleteInstanceNotIn(nil)

	require.Error(t, err)
}

func Test_NewThemeRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewThemeRepository(db, "theme")

	require.NoError(t, err)
	require.NotNil(t, repo)
}

func Test_NullableString(t *testing.T) {
	require.Nil(t, nullableString(""))
	require.Equal(t, "value", nullableString("value"))
}
