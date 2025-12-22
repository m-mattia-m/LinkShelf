package repository

import (
	"backend/internal/infrastructure/api/model"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func Test_UserRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"email",
		"first_name",
		"last_name",
	}).AddRow(
		"user-uuid-test",
		"test@test.com",
		"First",
		"Last",
	)

	mock.ExpectQuery(`SELECT \* FROM "user"`).
		WillReturnRows(rows)

	users, err := repo.List()

	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, "test@test.com", users[0].Email)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectQuery(`SELECT \* FROM "user"`).
		WillReturnError(errors.New("query failed"))

	users, err := repo.List()

	require.Error(t, err)
	require.Nil(t, users)
}

func Test_UserRepository_Get_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id",
		"email",
		"first_name",
		"last_name",
		"password",
	}).AddRow(
		"user-uuid-test",
		"test@test.com",
		"First",
		"Last",
		"hashed-password",
	)

	mock.ExpectQuery(`SELECT \* FROM "user" WHERE id =`).
		WithArgs("user-uuid-test").
		WillReturnRows(rows)

	user, err := repo.Get("user-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "hashed-password", user.Password)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_Get_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectQuery(`SELECT \* FROM "user" WHERE id =`).
		WithArgs("user-uuid-test").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.Get("user-uuid-test")

	require.NoError(t, err)
	require.Nil(t, user)
}

func Test_UserRepository_GetPassword_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	rows := sqlmock.NewRows([]string{"password"}).
		AddRow("hashed-password")

	mock.ExpectQuery(`SELECT password FROM "user" WHERE id =`).
		WithArgs("user-uuid-test").
		WillReturnRows(rows)

	password, err := repo.GetPassword("user-uuid-test")

	require.NoError(t, err)
	require.Equal(t, "hashed-password", password)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_GetPassword_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectQuery(`SELECT password FROM "user" WHERE id =`).
		WithArgs("user-uuid-test").
		WillReturnError(sql.ErrNoRows)

	password, err := repo.GetPassword("user-uuid-test")

	require.NoError(t, err)
	require.Empty(t, password)
}

func Test_UserRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`INSERT INTO "user"`).
		WithArgs(
			sqlmock.AnyArg(), // generated UUID
			"test@test.com",
			"First",
			"Last",
			"hashed-password",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user := &model.User{
		UserBase: model.UserBase{
			Email:     "test@test.com",
			FirstName: "First",
			LastName:  "Last",
			Password:  "hashed-password",
		},
	}

	id, err := repo.Create(user)

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.Equal(t, id, user.Id)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_Create_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`INSERT INTO "user"`).
		WillReturnError(errors.New("insert failed"))

	id, err := repo.Create(&model.User{})

	require.Error(t, err)
	require.Empty(t, id)
}

func Test_UserRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`UPDATE "user"`).
		WithArgs(
			"new@test.com",
			"NewFirst",
			"NewLast",
			"user-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(&model.User{
		Id: "user-uuid-test",
		UserBase: model.UserBase{
			Email:     "new@test.com",
			FirstName: "NewFirst",
			LastName:  "NewLast",
		},
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_Update_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`UPDATE "user"`).
		WillReturnError(errors.New("update failed"))

	err = repo.Update(&model.User{Id: "user-uuid-test"})

	require.Error(t, err)
}

func Test_UserRepository_PatchPassword_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`UPDATE "user" SET password`).
		WithArgs(
			"new-password",
			"user-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.PatchPassword(&model.User{
		Id: "user-uuid-test",
		UserBase: model.UserBase{
			Password: "new-password",
		},
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_PatchPassword_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`UPDATE "user" SET password`).
		WillReturnError(errors.New("patch failed"))

	err = repo.PatchPassword(&model.User{Id: "user-uuid-test"})

	require.Error(t, err)
}

func Test_UserRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`DELETE FROM "user"`).
		WithArgs("user-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(&model.User{Id: "user-uuid-test"})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_Delete_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`DELETE FROM "user"`).
		WillReturnError(errors.New("delete failed"))

	err = repo.Delete(&model.User{Id: "user-uuid-test"})

	require.Error(t, err)
}
