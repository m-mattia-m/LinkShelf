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
		"no_index",
		"footer_enabled",
		"footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at",
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
		false,
		true,
		"",
		"", "", "", "", 0,
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
		"no_index",
		"footer_enabled",
		"footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at",
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
		false,
		true,
		"",
		"", "", "", "", 0,
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
			false,
			false,
			"",
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
			"updated.com", "updated.com", "updated.com", "updated.com", "updated.com", sqlmock.AnyArg(),
			"updated-title",
			"/updated",
			"updated.com",
			"updated-desc",
			"light",
			"updated-icon",
			false,
			false,
			"",
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
		"no_index",
		"footer_enabled",
		"footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at",
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
		false,
		true,
		"",
		"", "", "", "", 0,
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
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths", "no_index", "footer_enabled", "footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at",
	}).AddRow(
		"shelf-uuid-test", "test-shelf", "/test", "example.com", "description-test", "dark", "icon-test", "user-uuid-test", "owner-name", false, false, true, "",
		"", "", "", "", 0,
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
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths", "no_index", "footer_enabled", "footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at",
	}).AddRow(
		"shelf-uuid-test", "test-shelf", "my-path", nil, "", nil, "", "user-uuid-test", "owner-name", false, false, true, "",
		"", "", "", "", 0,
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
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths", "no_index", "footer_enabled", "footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at",
	}).AddRow(
		"shelf-uuid-test", "test-shelf", "my-path", nil, "", nil, "", "user-uuid-test", "alice", false, false, true, "",
		"", "", "", "", 0,
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
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths", "no_index", "footer_enabled", "footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at",
	}).
		AddRow("shelf-1", "Old", "old-path", nil, "", nil, "", "user-1", "alice", false, false, true, "", "", "", "", "", 0).
		AddRow("shelf-2", "New", "new-path", nil, "", nil, "", "user-1", "alice", true, false, true, "", "", "", "", "", 0)
	mock.ExpectQuery(`FROM\s+shelf`).WillReturnRows(rows)

	shelves, err := repo.List()

	require.NoError(t, err)
	require.False(t, shelves[0].CreatedWithUserBasedPaths)
	require.True(t, shelves[1].CreatedWithUserBasedPaths)

	mock.ExpectExec("INSERT INTO shelf").
		WithArgs(sqlmock.AnyArg(), "Created", nil, nil, "", nil, "", "user-1", true, false, false, "").
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
	mock.ExpectExec(`^\s*UPDATE shelf\s+SET domain_state = CASE.*,\s+title = \$7,\s+path = \$8,\s+domain = \$9,\s+description = \$10,\s+theme_id = \$11,\s+icon = \$12,\s+no_index = \$13,\s+footer_enabled = \$14,\s+footer_custom_text = \$15\s+WHERE id = \$16\s*$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(&model.Shelf{
		PublicShelf:               model.PublicShelf{Id: "shelf-1", Title: "T"},
		CreatedWithUserBasedPaths: true,
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_GetByDomain_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths", "no_index", "footer_enabled", "footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at",
	}).AddRow(
		"shelf-uuid-test", "test-shelf", nil, "profile.example.com", "description-test", "", "icon-test", "user-uuid-test", "owner-name", false, false, true, "",
		"", "", "", "", 0,
	)

	mock.ExpectQuery(`(?s)FROM\s+shelf s.*WHERE s\.domain =`).
		WithArgs("profile.example.com").
		WillReturnRows(rows)

	shelf, err := repo.GetByDomain("profile.example.com")

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "profile.example.com", shelf.Domain)
	require.Empty(t, shelf.Path, "a domain-only shelf has a NULL path")

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_GetByDomain_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("nobody.example.com").
		WillReturnError(sql.ErrNoRows)

	shelf, err := repo.GetByDomain("nobody.example.com")

	require.NoError(t, err)
	require.Nil(t, shelf)
}

func Test_ShelfRepository_GetByDomain_Failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+shelf`).
		WithArgs("profile.example.com").
		WillReturnError(errors.New("db unavailable"))

	shelf, err := repo.GetByDomain("profile.example.com")

	require.ErrorContains(t, err, "db unavailable")
	require.Nil(t, shelf)
}

func Test_ShelfRepository_DomainInUse(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`(?s)SELECT COUNT\(\*\).*FROM\s+shelf.*WHERE domain = .* AND id <>`).
		WithArgs("profile.example.com", "shelf-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	inUse, err := repo.DomainInUse("profile.example.com", "shelf-1")
	require.NoError(t, err)
	require.True(t, inUse)

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs("free.example.com", "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	inUse, err = repo.DomainInUse("free.example.com", "")
	require.NoError(t, err)
	require.False(t, inUse)

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs("broken.example.com", "").
		WillReturnError(errors.New("db unavailable"))
	_, err = repo.DomainInUse("broken.example.com", "")
	require.ErrorContains(t, err, "db unavailable")

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_CountByUserId(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`SELECT COUNT\(\*\)\s+FROM shelf\s+WHERE user_id =`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := repo.CountByUserId("user-1")

	require.NoError(t, err)
	require.Equal(t, 3, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_Update_ResetsTheStatusOnlyWhenTheDomainChanges(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	// The assignments of the status come before domain = ?, so MySQL (which evaluates
	// them in order) still compares against the old domain.
	mock.ExpectExec(`SET domain_state = CASE WHEN COALESCE\(domain, ''\) = \$1 THEN domain_state ELSE '' END.*domain_updated_at = CASE.*ELSE \$6 END.*domain = \$9`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.Update(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-1", Title: "T"}, Domain: "profile.example.com"}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_Scan_ReadsTheDomainStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	columns := []string{"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths", "no_index", "footer_enabled", "footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at"}
	records := `[{"type":"TXT","name":"_c.profile.example.com","value":"v"}]`
	mock.ExpectQuery(`FROM\s+shelf`).WillReturnRows(sqlmock.NewRows(columns).
		AddRow("s-1", "T", nil, "profile.example.com", "", nil, "", "u-1", "alice", false, false, true, "", "pending", "txt_missing", "m", records, 1700000000).
		AddRow("s-2", "T", nil, "other.example.com", "", nil, "", "u-1", "alice", false, false, true, "", "", "", "", "", 0).
		AddRow("s-3", "T", nil, "broken.example.com", "", nil, "", "u-1", "alice", false, false, true, "", "error", "", "", "{not json", 0))

	shelves, err := repo.List()
	require.NoError(t, err)

	require.Equal(t, &model.DomainStatus{State: "pending", Reason: "txt_missing", Message: "m",
		Records: []model.DomainRecord{{Type: "TXT", Name: "_c.profile.example.com", Value: "v"}}}, shelves[0].DomainStatus)
	require.EqualValues(t, 1700000000, shelves[0].DomainUpdatedAt)
	require.Nil(t, shelves[1].DomainStatus)
	// A damaged value costs the records, not the shelf.
	require.Equal(t, "error", shelves[2].DomainStatus.State)
	require.Empty(t, shelves[2].DomainStatus.Records)
}

func Test_ShelfRepository_ListDomains_AfterCursor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`WHERE s\.domain IS NOT NULL AND s\.id > \$1 ORDER BY s\.id LIMIT \$2`).
		WithArgs("cursor-id", 3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths", "no_index", "footer_enabled", "footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at"}).
			AddRow("s-1", "T", nil, "profile.example.com", "", nil, "", "u-1", "alice", false, false, true, "", "", "", "", "", 0))

	shelves, err := repo.ListDomains("cursor-id", 3)
	require.NoError(t, err)
	require.Len(t, shelves, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_ShelfRepository_SetDomainStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}
	status := model.DomainStatus{State: "pending", Reason: "txt_missing", Message: "m", Records: []model.DomainRecord{{Type: "TXT", Name: "n", Value: "v"}}}

	mock.ExpectExec(`UPDATE shelf\s+SET domain_state = \$1.*WHERE id = \$6 AND domain = \$7`).
		WithArgs("pending", "txt_missing", "m", `[{"type":"TXT","name":"n","value":"v"}]`, int64(5), "s-1", "profile.example.com").
		WillReturnResult(sqlmock.NewResult(0, 1))
	updated, err := repo.SetDomainStatus("s-1", "profile.example.com", status, 5)
	require.NoError(t, err)
	require.True(t, updated)

	// The shelf has another domain by now.
	mock.ExpectExec(`UPDATE shelf`).WillReturnResult(sqlmock.NewResult(0, 0))
	updated, err = repo.SetDomainStatus("s-1", "old.example.com", status, 5)
	require.NoError(t, err)
	require.False(t, updated)

	mock.ExpectExec(`UPDATE shelf`).WillReturnError(errors.New("boom"))
	_, err = repo.SetDomainStatus("s-1", "profile.example.com", status, 5)
	require.Error(t, err)
}

func Test_ShelfRepository_ReadyOnlyQueries(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &shelfRepository{Engine: db}

	mock.ExpectQuery(`WHERE s\.domain = \$1 AND s\.domain_state = 'ready'`).
		WithArgs("profile.example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "path", "domain", "description", "theme", "icon", "user_id", "username", "created_user_based_paths", "no_index", "footer_enabled", "footer_custom_text", "domain_state", "domain_reason", "domain_message", "domain_records", "domain_updated_at"}))
	shelf, err := repo.GetReadyByDomain("profile.example.com")
	require.NoError(t, err)
	require.Nil(t, shelf)

	mock.ExpectQuery(`domain = \$1 AND domain_state = 'ready' AND id <> \$2`).
		WithArgs("profile.example.com", "s-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	taken, err := repo.DomainReadyInUse("profile.example.com", "s-1")
	require.NoError(t, err)
	require.True(t, taken)
}
