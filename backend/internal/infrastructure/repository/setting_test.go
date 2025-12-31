//go:build integration
// +build integration

package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_SettingRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &settingRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"key",
		"value",
	}).AddRow(
		"theme",
		"dark",
	)

	mock.ExpectQuery(`FROM\s+setting`).
		WillReturnRows(rows)

	settings, err := repo.List()

	require.NoError(t, err)
	require.Len(t, settings, 1)
	require.Equal(t, "theme", settings[0].Key)
	require.Equal(t, "dark", settings[0].Value)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SettingRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &settingRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+setting`).
		WillReturnError(errors.New("query failed"))

	settings, err := repo.List()

	require.Error(t, err)
	require.Nil(t, settings)
}

func Test_SettingRepository_List_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &settingRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"key"}).
		AddRow("only-key")

	mock.ExpectQuery(`FROM\s+setting`).
		WillReturnRows(rows)

	settings, err := repo.List()

	require.Error(t, err)
	require.Nil(t, settings)
}

func Test_SettingRepository_GetByKey_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &settingRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"key",
		"value",
	}).AddRow(
		"theme",
		"dark",
	)

	mock.ExpectQuery(`FROM\s+setting`).
		WithArgs("theme").
		WillReturnRows(rows)

	setting, err := repo.GetByKey("theme")

	require.NoError(t, err)
	require.NotNil(t, setting)
	require.Equal(t, "theme", setting.Key)
	require.Equal(t, "dark", setting.Value)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SettingRepository_GetByKey_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &settingRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+setting`).
		WithArgs("missing").
		WillReturnError(sqlmock.ErrCancelled) // forces QueryRow failure

	setting, err := repo.GetByKey("missing")

	require.Error(t, err)
	require.Nil(t, setting)
}

func Test_SettingRepository_GetByKey_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &settingRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+setting`).
		WithArgs("theme").
		WillReturnError(errors.New("query failed"))

	setting, err := repo.GetByKey("theme")

	require.Error(t, err)
	require.Nil(t, setting)
}

func Test_SettingRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &settingRepository{Engine: db}

	mock.ExpectExec(`UPDATE\s+setting`).
		WithArgs(
			"dark",
			"theme",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update("theme", "dark")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SettingRepository_Update_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &settingRepository{Engine: db}

	mock.ExpectExec(`UPDATE\s+setting`).
		WillReturnError(errors.New("update failed"))

	err = repo.Update("theme", "dark")

	require.Error(t, err)
}
