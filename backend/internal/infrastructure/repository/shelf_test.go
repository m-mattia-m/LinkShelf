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
		"username",
		"created_user_based_paths",
	}).AddRow(
		"shelf-uuid-test",
		"test-shelf",
		"/test",
		"example.com",
		"description-test",
		"dark",
		"icon-test",
		"user-uuid-test",
		"owner-name",
		false,
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
		"username",
		"created_user_based_paths",
	}).AddRow(
		"shelf-uuid-test",
		"test-shelf",
		"/test",
		"example.com",
		"description-test",
		"dark",
		"icon-test",
		"user-uuid-test",
		"owner-name",
		false,
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
			false,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	shelf := &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       "test-shelf",
			Path:        "/test",
			Description: "description-test",
			Icon:        "icon-test",
		},
		Domain:  "example.com",
		ThemeId: "dark",
		UserId:  "user-uuid-test",
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
		Domain:  "updated.com",
		ThemeId: "light",
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
		"username",
		"created_user_based_paths",
	}).AddRow(
		"shelf-uuid-test",
		"test-shelf",
		"my-path",
		"",
		"description-test",
		"",
		"icon-test",
		"user-uuid-test",
		"owner-name",
		false,
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

func Test_ShelfRepository_ListByUserId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths",
	}).AddRow(
		"shelf-uuid-test", "test-shelf", "/test", "example.com", "description-test", "dark", "icon-test", "user-uuid-test", "owner-name", false,
	)

	mock.ExpectQuery(`(?s)FROM\s+shelf s.*WHERE s\.user_id =`).
		WithArgs("user-uuid-test").
		WillReturnRows(rows)

	shelves, err := repo.ListByUserId("user-uuid-test")

	require.NoError(t, err)
	require.Len(t, shelves, 1)
	require.Equal(t, "user-uuid-test", shelves[0].UserId)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_ListByUserId_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`(?s)FROM\s+shelf s.*WHERE s\.user_id =`).
		WillReturnError(errors.New("query failed"))

	shelves, err := repo.ListByUserId("user-uuid-test")

	require.Error(t, err)
	require.Nil(t, shelves)
}

func Test_ShelfRepository_List_FillsOwnerUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths",
	}).AddRow(
		"shelf-uuid-test", "test-shelf", "my-path", nil, "", nil, "", "user-uuid-test", "owner-name", false,
	)
	mock.ExpectQuery(`(?s)FROM\s+shelf s\s+JOIN\s+"user" u ON u\.id = s\.user_id`).WillReturnRows(rows)

	shelves, err := repo.List()

	require.NoError(t, err)
	require.Len(t, shelves, 1)
	require.Equal(t, "owner-name", shelves[0].Username)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_GetByUsernameAndPath_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths",
	}).AddRow(
		"shelf-uuid-test", "test-shelf", "my-path", nil, "", nil, "", "user-uuid-test", "alice", false,
	)
	mock.ExpectQuery(`(?s)WHERE u\.username = .* AND LOWER\(s\.path\) = LOWER`).
		WithArgs("alice", "my-path").
		WillReturnRows(rows)

	shelf, err := repo.GetByUsernameAndPath("alice", "my-path")

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "alice", shelf.Username)
	require.Equal(t, "my-path", shelf.Path)
}

func Test_ShelfRepository_GetByUsernameAndPath_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).WithArgs("alice", "missing").WillReturnError(sql.ErrNoRows)

	shelf, err := repo.GetByUsernameAndPath("alice", "missing")

	require.NoError(t, err)
	require.Nil(t, shelf)
}

func Test_ShelfRepository_PathInUse(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`(?s)FROM\s+shelf\s+WHERE LOWER\(path\) = LOWER\(.*\) AND id <>`).
		WithArgs("my-path", "shelf-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("free", "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	taken, err := repo.PathInUse("my-path", "shelf-1")
	require.NoError(t, err)
	require.True(t, taken)

	taken, err = repo.PathInUse("free", "")
	require.NoError(t, err)
	require.False(t, taken)
}

func Test_ShelfRepository_PathInUseByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`(?s)WHERE user_id = .* AND LOWER\(path\)`).
		WithArgs("user-1", "my-path", "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	taken, err := repo.PathInUseByUser("user-1", "my-path", "")

	require.NoError(t, err)
	require.True(t, taken)
}

func Test_ShelfRepository_PathInUse_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).WillReturnError(errors.New("boom"))

	_, err = repo.PathInUse("p", "")
	require.Error(t, err)

	mock.ExpectQuery(`FROM\s+shelf`).WillReturnError(errors.New("boom"))

	_, err = repo.PathInUseByUser("u", "p", "")
	require.Error(t, err)
}

func Test_ShelfRepository_ListPathCollisions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"path", "id", "user_id", "username"}).
		AddRow("profile", "shelf-1", "user-1", "alice").
		AddRow("Profile", "shelf-2", "user-2", nil)
	mock.ExpectQuery(`(?s)GROUP BY lower_path\s+HAVING COUNT\(\*\) > 1`).WillReturnRows(rows)

	collisions, err := repo.ListPathCollisions()

	require.NoError(t, err)
	require.Equal(t, []PathCollision{
		{Path: "profile", ShelfId: "shelf-1", UserId: "user-1", Username: "alice"},
		{Path: "Profile", ShelfId: "shelf-2", UserId: "user-2"},
	}, collisions)
}

func Test_ShelfRepository_ListPathCollisions_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).WillReturnError(errors.New("boom"))

	_, err = repo.ListPathCollisions()
	require.Error(t, err)
}

func Test_ShelfRepository_ReadsAndStoresTheCreationMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths",
	}).
		AddRow("shelf-1", "Old", "old-path", nil, "", nil, "", "user-1", "alice", false).
		AddRow("shelf-2", "New", "new-path", nil, "", nil, "", "user-1", "alice", true)
	mock.ExpectQuery(`FROM\s+shelf`).WillReturnRows(rows)

	shelves, err := repo.List()

	require.NoError(t, err)
	require.False(t, shelves[0].CreatedWithUserBasedPaths)
	require.True(t, shelves[1].CreatedWithUserBasedPaths)

	mock.ExpectExec("INSERT INTO shelf").
		WithArgs(sqlmock.AnyArg(), "Created", nil, nil, "", nil, "", "user-1", true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = repo.Create(&model.Shelf{
		PublicShelf:               model.PublicShelf{Title: "Created"},
		UserId:                    "user-1",
		CreatedWithUserBasedPaths: true,
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_UpdateNeverTouchesTheCreationMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	// The UPDATE statement must not mention the column at all.
	mock.ExpectExec(`^\s*UPDATE shelf\s+SET title = \$1,\s+path = \$2,\s+domain = \$3,\s+description = \$4,\s+theme_id = \$5,\s+icon = \$6\s+WHERE id = \$7\s*$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(&model.Shelf{
		PublicShelf:               model.PublicShelf{Id: "shelf-1", Title: "T"},
		CreatedWithUserBasedPaths: true,
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
