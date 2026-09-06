package repository

import (
	"backend/internal/infrastructure/api/model"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_ShelfRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"path",
		"domain",
		"description",
		"theme",
		"icon",
		"user_id",
	}).AddRow(
		"shelf-uuid-test",
		"test-shelf",
		"/test",
		"example.com",
		"description-test",
		"dark",
		"icon-test",
		"user-uuid-test",
	)

	mock.ExpectQuery(`FROM\s+shelf`).
		WillReturnRows(rows)

	shelves, err := repo.List()

	require.NoError(t, err)
	require.Len(t, shelves, 1)
	require.Equal(t, "shelf-uuid-test", shelves[0].Id)
	require.Equal(t, "test-shelf", shelves[0].Title)
	require.Equal(t, "/test", shelves[0].Path)
	require.Equal(t, "example.com", shelves[0].Domain)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_List_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).
		WillReturnError(sql.ErrNoRows)

	shelves, err := repo.List()

	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, shelves)

	require.NoError(t, mock.ExpectationsWereMet())
}

//func Test_ShelfRepository_List_ScanError(t *testing.T) {
//	db, mock, err := sqlmock.New()
//	require.NoError(t, err)
//	defer db.Close()
//
//	repo := &shelfRepository{Engine: db}
//
//	// Force type mismatch: string into int field
//	rows := sqlmock.NewRows([]string{
//		"id",
//		"title",
//		"description",
//		"theme",
//		"icon",
//		"user_id",
//	}).AddRow(
//		123,
//		"test-shelf",
//		"description",
//		"dark",
//		"icon",
//		"user-id",
//	)
//
//	mock.ExpectQuery(`FROM\s+shelf`).
//		WillReturnRows(rows)
//
//	shelf, err := repo.List()
//
//	require.Error(t, err)
//	require.Nil(t, shelf)
//}

func Test_ShelfRepository_Get_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"path",
		"domain",
		"description",
		"theme",
		"icon",
		"user_id",
	}).AddRow(
		"shelf-uuid-test",
		"test-shelf",
		"/test",
		"example.com",
		"description-test",
		"dark",
		"icon-test",
		"user-uuid-test",
	)

	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("shelf-uuid-test").
		WillReturnRows(rows)

	shelf, err := repo.Get("shelf-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "/test", shelf.Path)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_Get_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("shelf-uuid-test").
		WillReturnError(sql.ErrNoRows)

	shelf, err := repo.Get("shelf-uuid-test")

	require.NoError(t, err)
	require.Nil(t, shelf)
}

//func Test_ShelfRepository_Get_QueryError(t *testing.T) {
//	db, mock, err := sqlmock.New()
//	require.NoError(t, err)
//	defer db.Close()
//
//	repo := &shelfRepository{Engine: db}
//
//	// Match only the FROM clause (stable)
//	mock.ExpectQuery(`FROM\s+shelf`).
//		WithArgs("shelf-uuid-test").
//		WillReturnError(errors.New("query failed"))
//
//	shelf, err := repo.Get("shelf-uuid-test")
//
//	require.Error(t, err)
//	require.Nil(t, shelf)
//	require.NoError(t, mock.ExpectationsWereMet())
//}

func Test_ShelfRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectExec("INSERT INTO shelf").
		WithArgs(
			sqlmock.AnyArg(), // generated UUID
			"test-shelf",
			"/test",
			"example.com",
			"description-test",
			"dark",
			"icon-test",
			"user-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	shelf := &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       "test-shelf",
			Path:        "/test",
			Description: "description-test",
			Icon:        "icon-test",
		},
		Domain: "example.com",
		Theme:  "dark",
		UserId: "user-uuid-test",
	}

	id, err := repo.Create(shelf)

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.Equal(t, id, shelf.Id)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectExec("UPDATE shelf").
		WithArgs(
			"updated-title",
			"/updated",
			"updated.com",
			"updated-desc",
			"light",
			"updated-icon",
			"shelf-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(&model.Shelf{
		PublicShelf: model.PublicShelf{
			Id:          "shelf-uuid-test",
			Title:       "updated-title",
			Path:        "/updated",
			Description: "updated-desc",
			Icon:        "updated-icon",
		},
		Domain: "updated.com",
		Theme:  "light",
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_Update_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectExec("UPDATE shelf").
		WillReturnError(errors.New("update failed"))

	err = repo.Update(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}})

	require.Error(t, err)
}

func Test_ShelfRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectExec("DELETE FROM shelf").
		WithArgs("shelf-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_Delete_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectExec("DELETE FROM shelf").
		WillReturnError(errors.New("delete failed"))

	err = repo.Delete(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}})

	require.Error(t, err)
}

func Test_ShelfRepository_GetByPath_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"path",
		"domain",
		"description",
		"theme",
		"icon",
		"user_id",
	}).AddRow(
		"shelf-uuid-test",
		"test-shelf",
		"my-path",
		"",
		"description-test",
		"",
		"icon-test",
		"user-uuid-test",
	)

	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("my-path").
		WillReturnRows(rows)

	shelf, err := repo.GetByPath("my-path")

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "my-path", shelf.Path)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_GetByPath_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("missing-path").
		WillReturnError(sql.ErrNoRows)

	shelf, err := repo.GetByPath("missing-path")

	require.NoError(t, err)
	require.Nil(t, shelf)
}
