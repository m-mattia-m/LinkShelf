package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_OidcStateRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &oidcStateRepository{Engine: db}

	mock.ExpectExec(`INSERT INTO "oidc_state"`).
		WithArgs("state-test", "verifier-test", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create("state-test", "verifier-test", time.Now().Add(10*time.Minute))

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_OidcStateRepository_GetByState_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &oidcStateRepository{Engine: db}

	expiry := time.Now().Add(10 * time.Minute)
	rows := sqlmock.NewRows([]string{"state", "code_verifier", "expires_at"}).
		AddRow("state-test", "verifier-test", expiry)

	mock.ExpectQuery(`FROM\s+"oidc_state"\s+WHERE state =`).
		WithArgs("state-test").
		WillReturnRows(rows)

	result, err := repo.GetByState("state-test")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "verifier-test", result.CodeVerifier)
}

func Test_OidcStateRepository_GetByState_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &oidcStateRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+"oidc_state"\s+WHERE state =`).
		WithArgs("missing-state").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByState("missing-state")

	require.NoError(t, err)
	require.Nil(t, result)
}

func Test_OidcStateRepository_DeleteByState_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &oidcStateRepository{Engine: db}

	mock.ExpectExec(`DELETE FROM "oidc_state"`).
		WithArgs("state-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.DeleteByState("state-test")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_OidcStateRepository_DeleteByState_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &oidcStateRepository{Engine: db}

	mock.ExpectExec(`DELETE FROM "oidc_state"`).
		WillReturnError(errors.New("delete failed"))

	err = repo.DeleteByState("state-test")

	require.Error(t, err)
}
