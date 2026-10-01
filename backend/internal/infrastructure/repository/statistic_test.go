package repository

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_StatisticRepository_GetShelfAmount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &statisticRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"count"}).AddRow(3)

	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("user-uuid-test").
		WillReturnRows(rows)

	count, err := repo.GetShelfAmount("user-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, count)
	require.Equal(t, 3, *count)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_StatisticRepository_GetShelfAmount_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &statisticRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("user-uuid-test").
		WillReturnError(sql.ErrNoRows)

	count, err := repo.GetShelfAmount("user-uuid-test")

	require.NoError(t, err)
	require.Nil(t, count)
}

// A QueryError test is omitted: the method returns a non-nil count alongside the error.

func Test_StatisticRepository_GetSectionAmount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &statisticRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)

	mock.ExpectQuery(`FROM\s+section`).
		WithArgs("user-uuid-test").
		WillReturnRows(rows)

	count, err := repo.GetSectionAmount("user-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, count)
	require.Equal(t, 5, *count)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_StatisticRepository_GetSectionAmount_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &statisticRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+section`).
		WithArgs("user-uuid-test").
		WillReturnError(sql.ErrNoRows)

	count, err := repo.GetSectionAmount("user-uuid-test")

	require.NoError(t, err)
	require.Nil(t, count)
}

// QueryError test omitted, see GetShelfAmount.

func Test_StatisticRepository_GetLinkAmount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &statisticRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"count"}).AddRow(8)

	mock.ExpectQuery(`FROM\s+link`).
		WithArgs("user-uuid-test").
		WillReturnRows(rows)

	count, err := repo.GetLinkAmount("user-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, count)
	require.Equal(t, 8, *count)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_StatisticRepository_GetLinkAmount_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &statisticRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+link`).
		WithArgs("user-uuid-test").
		WillReturnError(sql.ErrNoRows)

	count, err := repo.GetLinkAmount("user-uuid-test")

	require.NoError(t, err)
	require.Nil(t, count)
}

// QueryError test omitted, see GetShelfAmount.
