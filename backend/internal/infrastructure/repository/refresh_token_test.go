package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_RefreshTokenRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	mock.ExpectExec(`INSERT INTO refresh_token`).
		WithArgs(sqlmock.AnyArg(), "user-uuid-test", "hash-test", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create("user-uuid-test", "hash-test", time.Now().Add(time.Hour))

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_RefreshTokenRepository_Create_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	mock.ExpectExec(`INSERT INTO refresh_token`).
		WillReturnError(errors.New("insert failed"))

	err = repo.Create("user-uuid-test", "hash-test", time.Now().Add(time.Hour))

	require.Error(t, err)
}

func Test_RefreshTokenRepository_GetByHash_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	expiry := time.Now().Add(time.Hour)
	rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at"}).
		AddRow("token-uuid-test", "user-uuid-test", "hash-test", expiry)

	mock.ExpectQuery(`FROM\s+refresh_token\s+WHERE token_hash =`).
		WithArgs("hash-test").
		WillReturnRows(rows)

	result, err := repo.GetByHash("hash-test")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "user-uuid-test", result.UserId)
	require.Equal(t, "hash-test", result.TokenHash)
}

func Test_RefreshTokenRepository_GetByHash_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+refresh_token\s+WHERE token_hash =`).
		WithArgs("missing-hash").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByHash("missing-hash")

	require.NoError(t, err)
	require.Nil(t, result)
}

func Test_RefreshTokenRepository_GetByHash_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+refresh_token\s+WHERE token_hash =`).
		WillReturnError(errors.New("query failed"))

	result, err := repo.GetByHash("hash-test")

	require.Error(t, err)
	require.Nil(t, result)
}

func Test_RefreshTokenRepository_DeleteByHash_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	mock.ExpectExec(`DELETE FROM refresh_token\s+WHERE token_hash =`).
		WithArgs("hash-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.DeleteByHash("hash-test")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_RefreshTokenRepository_DeleteByHash_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	mock.ExpectExec(`DELETE FROM refresh_token\s+WHERE token_hash =`).
		WillReturnError(errors.New("delete failed"))

	err = repo.DeleteByHash("hash-test")

	require.Error(t, err)
}

func Test_RefreshTokenRepository_DeleteByUserId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	mock.ExpectExec(`DELETE FROM refresh_token\s+WHERE user_id =`).
		WithArgs("user-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 2))

	err = repo.DeleteByUserId("user-uuid-test")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_RefreshTokenRepository_DeleteByUserId_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &refreshTokenRepository{Engine: db}

	mock.ExpectExec(`DELETE FROM refresh_token\s+WHERE user_id =`).
		WillReturnError(errors.New("delete failed"))

	err = repo.DeleteByUserId("user-uuid-test")

	require.Error(t, err)
}
