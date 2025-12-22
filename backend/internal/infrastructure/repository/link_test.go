package repository

import (
	"backend/internal/infrastructure/api/model"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_LinkRepository_ListByShelfId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{
		Engine: db,
		Table:  "link",
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"link",
		"icon",
		"color",
		"section_id",
	}).AddRow(
		"link-uuid-test",
		"title-test",
		"https://example.com",
		"icon-test",
		"#ff0000",
		"section-uuid-test",
	)

	mock.ExpectQuery("SELECT l.\\*").
		WithArgs("shelf-uuid-test").
		WillReturnRows(rows)

	links, err := repo.ListByShelfId("shelf-uuid-test")

	require.NoError(t, err)
	require.Len(t, links, 1)

	require.Equal(t, "link-uuid-test", links[0].Id)
	require.Equal(t, "title-test", links[0].Title)
	require.Equal(t, "#ff0000", links[0].Color)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_LinkRepository_ListByShelfId_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	mock.ExpectQuery("SELECT l.\\*").
		WithArgs("shelf-uuid-test").
		WillReturnError(errors.New("db error"))

	links, err := repo.ListByShelfId("shelf-uuid-test")

	require.Error(t, err)
	require.Nil(t, links)
}

func Test_LinkRepository_ListByShelfId_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
	}).AddRow("id-only", "title-only")

	mock.ExpectQuery("SELECT l.\\*").
		WithArgs("shelf-uuid-test").
		WillReturnRows(rows)

	links, err := repo.ListByShelfId("shelf-uuid-test")

	require.Error(t, err)
	require.Nil(t, links)
}

func Test_LinkRepository_Get_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"link",
		"icon",
		"color",
		"section_id",
	}).AddRow(
		"link-uuid-test",
		"title-test",
		"https://example.com",
		"icon-test",
		"#ff0000",
		"section-uuid-test",
	)

	mock.ExpectQuery("SELECT \\* FROM link").
		WithArgs("link-uuid-test").
		WillReturnRows(rows)

	link, err := repo.Get("link-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, link)
	require.Equal(t, "link-uuid-test", link.Id)
	require.Equal(t, "#ff0000", link.Color)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_LinkRepository_Get_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	mock.ExpectQuery("SELECT \\* FROM link").
		WithArgs("link-uuid-test").
		WillReturnError(sql.ErrNoRows)

	link, err := repo.Get("link-uuid-test")

	require.NoError(t, err)
	require.Nil(t, link)
}

func Test_LinkRepository_Get_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow("only-id")

	mock.ExpectQuery("SELECT \\* FROM link").
		WithArgs("link-uuid-test").
		WillReturnRows(rows)

	link, err := repo.Get("link-uuid-test")

	require.Error(t, err)
	require.Nil(t, link)
}

func Test_LinkRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	mock.ExpectExec("INSERT INTO link").
		WithArgs(
			sqlmock.AnyArg(), // generated UUID
			"title-test",
			"https://example.com",
			"icon-test",
			"#ff0000",
			"section-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	link := &model.Link{
		LinkBase: model.LinkBase{
			Title:     "title-test",
			Link:      "https://example.com",
			Icon:      "icon-test",
			Color:     "#ff0000",
			SectionId: "section-uuid-test",
		},
	}

	id, err := repo.Create(link)

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.Equal(t, id, link.Id)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_LinkRepository_Create_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	mock.ExpectExec("INSERT INTO link").
		WillReturnError(errors.New("insert failed"))

	link := &model.Link{}

	id, err := repo.Create(link)

	require.Error(t, err)
	require.Empty(t, id)
}

func Test_LinkRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	mock.ExpectExec("UPDATE link").
		WithArgs(
			"title-updated",
			"https://example.com",
			"icon-test",
			"#00ff00",
			"link-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(&model.Link{
		Id: "link-uuid-test",
		LinkBase: model.LinkBase{
			Title: "title-updated",
			Link:  "https://example.com",
			Icon:  "icon-test",
			Color: "#00ff00",
		},
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_LinkRepository_Update_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	mock.ExpectExec("UPDATE link").
		WillReturnError(errors.New("update failed"))

	err = repo.Update(&model.Link{Id: "link-uuid-test"})

	require.Error(t, err)
}

func Test_LinkRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	mock.ExpectExec("DELETE FROM link").
		WithArgs("link-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(&model.Link{Id: "link-uuid-test"})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_LinkRepository_Delete_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &linkRepository{Engine: db}

	mock.ExpectExec("DELETE FROM link").
		WillReturnError(errors.New("delete failed"))

	err = repo.Delete(&model.Link{Id: "link-uuid-test"})

	require.Error(t, err)
}
