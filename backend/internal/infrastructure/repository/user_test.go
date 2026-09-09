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
		"role",
	}).AddRow(
		"user-uuid-test",
		"test@test.com",
		"First",
		"Last",
		"user",
	)

	mock.ExpectQuery(`FROM\s+"user"`).
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

	mock.ExpectQuery(`FROM\s+"user"`).
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
		"role",
	}).AddRow(
		"user-uuid-test",
		"test@test.com",
		"First",
		"Last",
		"user",
	)

	mock.ExpectQuery(`FROM\s+"user"\s+WHERE id =`).
		WithArgs("user-uuid-test").
		WillReturnRows(rows)

	user, err := repo.Get("user-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "test@test.com", user.Email)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_Get_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+"user"\s+WHERE id =`).
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
			"user",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	base := model.UserBase{
		Email:     "test@test.com",
		FirstName: "First",
		LastName:  "Last",
	}

	id, err := repo.Create(base, "hashed-password", "user")

	require.NoError(t, err)
	require.NotEmpty(t, id)

	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_Create_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`INSERT INTO "user"`).
		WillReturnError(errors.New("insert failed"))

	id, err := repo.Create(model.UserBase{}, "hashed-password", "user")

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
			"admin",
			"user-uuid-test",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(&model.User{
		Id: "user-uuid-test",
		UserBase: model.UserBase{
			Email:     "new@test.com",
			FirstName: "NewFirst",
			LastName:  "NewLast",
			Role:      "admin",
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

	err = repo.PatchPassword("user-uuid-test", "new-password")

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

	err = repo.PatchPassword("user-uuid-test", "new-password")

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

func Test_UserRepository_FindByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	rows := sqlmock.NewRows([]string{
		"id", "email", "first_name", "last_name", "role", "password", "provider", "provider_id",
	}).AddRow(
		"user-uuid-test", "test@test.com", "First", "Last", "user", "hashed", "LOCAL", nil,
	)

	mock.ExpectQuery(`FROM\s+"user"\s+WHERE LOWER\(email\)`).
		WithArgs("test@test.com").
		WillReturnRows(rows)

	record, err := repo.FindByEmail("test@test.com")

	require.NoError(t, err)
	require.NotNil(t, record)
	require.Equal(t, "user-uuid-test", record.Id)
	require.Equal(t, "LOCAL", record.Provider)
	require.Nil(t, record.ProviderId)
}

func Test_UserRepository_FindByEmail_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+"user"\s+WHERE LOWER\(email\)`).
		WithArgs("missing@test.com").
		WillReturnError(sql.ErrNoRows)

	record, err := repo.FindByEmail("missing@test.com")

	require.NoError(t, err)
	require.Nil(t, record)
}

func Test_UserRepository_FindByEmail_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+"user"\s+WHERE LOWER\(email\)`).
		WillReturnError(errors.New("query failed"))

	record, err := repo.FindByEmail("test@test.com")

	require.Error(t, err)
	require.Nil(t, record)
}

func Test_UserRepository_FindByProviderId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	providerId := "oidc-subject-test"
	rows := sqlmock.NewRows([]string{
		"id", "email", "first_name", "last_name", "role", "password", "provider", "provider_id",
	}).AddRow(
		"user-uuid-test", "test@test.com", "First", "Last", "user", "", "OIDC", providerId,
	)

	mock.ExpectQuery(`FROM\s+"user"\s+WHERE provider_id =`).
		WithArgs(providerId).
		WillReturnRows(rows)

	record, err := repo.FindByProviderId(providerId)

	require.NoError(t, err)
	require.NotNil(t, record)
	require.Equal(t, "OIDC", record.Provider)
	require.NotNil(t, record.ProviderId)
	require.Equal(t, providerId, *record.ProviderId)
}

func Test_UserRepository_FindByProviderId_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+"user"\s+WHERE provider_id =`).
		WithArgs("missing-subject").
		WillReturnError(sql.ErrNoRows)

	record, err := repo.FindByProviderId("missing-subject")

	require.NoError(t, err)
	require.Nil(t, record)
}

func Test_UserRepository_FindByProviderId_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectQuery(`FROM\s+"user"\s+WHERE provider_id =`).
		WillReturnError(errors.New("query failed"))

	record, err := repo.FindByProviderId("some-subject")

	require.Error(t, err)
	require.Nil(t, record)
}

func Test_UserRepository_CreateExternal_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), "test@test.com", "First", "Last", "OIDC", "oidc-subject-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := repo.CreateExternal("test@test.com", "First", "Last", "OIDC", "oidc-subject-test")

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_CreateExternal_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`INSERT INTO "user"`).
		WillReturnError(errors.New("insert failed"))

	_, err = repo.CreateExternal("test@test.com", "First", "Last", "OIDC", "oidc-subject-test")

	require.Error(t, err)
}

func Test_UserRepository_LinkProvider_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`UPDATE "user"\s+SET provider`).
		WithArgs("OIDC", "oidc-subject-test", "user-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.LinkProvider("user-uuid-test", "OIDC", "oidc-subject-test")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_LinkProvider_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`UPDATE "user"\s+SET provider`).
		WillReturnError(errors.New("update failed"))

	err = repo.LinkProvider("user-uuid-test", "OIDC", "oidc-subject-test")

	require.Error(t, err)
}

func Test_UserRepository_SetPasswordAndRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`UPDATE "user"\s+SET password`).
		WithArgs("new-hashed-password", "admin", "user-uuid-test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SetPasswordAndRole("user-uuid-test", "new-hashed-password", "admin")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_UserRepository_SetPasswordAndRole_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepository{Engine: db}

	mock.ExpectExec(`UPDATE "user"\s+SET password`).
		WillReturnError(errors.New("update failed"))

	err = repo.SetPasswordAndRole("user-uuid-test", "new-hashed-password", "admin")

	require.Error(t, err)
}
