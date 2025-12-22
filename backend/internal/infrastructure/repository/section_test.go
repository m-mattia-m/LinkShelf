package repository

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_SectionRepository_ListByShelfId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"shelf_id",
	}).AddRow(
		"section-uuid-test",
		"test-section",
		"shelf-uuid-test",
	)

	mock.ExpectQuery("SELECT \\* FROM section").
		WithArgs("shelf-uuid-test").
		WillReturnRows(rows)

	sections, err := repo.ListByShelfId("shelf-uuid-test")

	require.NoError(t, err)
	require.Len(t, sections, 1)
	require.Equal(t, "section-uuid-test", sections[0].Id)
	require.Equal(t, "test-section", sections[0].Title)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SectionRepository_ListByShelfId_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	mock.ExpectQuery("SELECT \\* FROM section").
		WithArgs("shelf-uuid-test").
		WillReturnError(errors.New("query failed"))

	sections, err := repo.ListByShelfId("shelf-uuid-test")

	require.Error(t, err)
	require.Nil(t, sections)
}

func Test_SectionRepository_ListByShelfId_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow("only-id")

	mock.ExpectQuery("SELECT \\* FROM section").
		WithArgs("shelf-uuid-test").
		WillReturnRows(rows)

	sections, err := repo.ListByShelfId("shelf-uuid-test")

	require.Error(t, err)
	require.Nil(t, sections)
}

func Test_SectionRepository_Get_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"shelf_id",
	}).AddRow(
		"section-uuid-test",
		"test-section",
		"shelf-uuid-test",
	)

	mock.ExpectQuery("SELECT \\* FROM section").
		WithArgs("section-uuid-test").
		WillReturnRows(rows)

	section, err := repo.Get("section-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, section)
	require.Equal(t, "test-section", section.Title)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SectionRepository_Get_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow("only-id")

	mock.ExpectQuery("SELECT \\* FROM section").
		WithArgs("section-uuid-test").
		WillReturnRows(rows)

	section, err := repo.Get("section-uuid-test")

	require.Error(t, err)
	require.Nil(t, section)
}

func Test_SectionRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	mock.ExpectExec("INSERT INTO section").
		WithArgs(
			sqlmock.AnyArg(), // generated UUID
			"test-section",
			"shelf-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	section := &model.Section{
		SectionBase: model.SectionBase{
			Title:   "test-section",
			ShelfId: "shelf-uuid-test",
		},
	}

	id, err := repo.Create(section)

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.Equal(t, id, section.Id)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SectionRepository_Create_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	mock.ExpectExec("INSERT INTO section").
		WillReturnError(errors.New("insert failed"))

	id, err := repo.Create(&model.Section{})

	require.Error(t, err)
	require.Empty(t, id)
}

func Test_SectionRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	mock.ExpectExec("UPDATE section").
		WithArgs(
			"updated-title",
			"section-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(&model.Section{
		Id: "section-uuid-test",
		SectionBase: model.SectionBase{
			Title: "updated-title",
		},
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SectionRepository_Update_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	mock.ExpectExec("UPDATE section").
		WillReturnError(errors.New("update failed"))

	err = repo.Update(&model.Section{Id: "section-uuid-test"})

	require.Error(t, err)
}

func Test_SectionRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	mock.ExpectExec("DELETE FROM section").
		WithArgs("section-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(&model.Section{Id: "section-uuid-test"})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SectionRepository_Delete_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &sectionRepository{Engine: db}

	mock.ExpectExec("DELETE FROM section").
		WillReturnError(errors.New("delete failed"))

	err = repo.Delete(&model.Section{Id: "section-uuid-test"})

	require.Error(t, err)
}
