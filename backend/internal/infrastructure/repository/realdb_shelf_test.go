//go:build realdb

package repository

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func realDBTestUser(t *testing.T) string {
	t.Helper()
	id, err := TestRepository.UserRepository.Create(model.UserBase{
		Email:     "realdb-shelf-" + uuid.NewString() + "@example.com",
		FirstName: "Shelf",
		LastName:  "Owner",
	}, "hashed", "user")
	require.NoError(t, err)
	return id
}

func Test_RealDB_ShelfRepository_CRUD(t *testing.T) {
	repo := TestRepository.ShelfRepository
	userId := realDBTestUser(t)

	id, err := repo.Create(&model.Shelf{
		PublicShelf: model.PublicShelf{Title: "My Shelf"},
		UserId:      userId,
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	fetched, err := repo.Get(id)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, "My Shelf", fetched.Title)
	require.Equal(t, "", fetched.Path, "an unset path must read back as empty string, not NULL/garbage")
	require.Equal(t, "", fetched.Domain)

	fetched.Title = "Renamed"
	fetched.Path = "renamed-path"
	require.NoError(t, repo.Update(fetched))

	updated, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, "Renamed", updated.Title)
	require.Equal(t, "renamed-path", updated.Path)

	byPath, err := repo.GetByPath("RENAMED-PATH")
	require.NoError(t, err, "GetByPath is documented as case-insensitive")
	require.NotNil(t, byPath)
	require.Equal(t, id, byPath.Id)

	shelves, err := repo.ListByUserId(userId)
	require.NoError(t, err)
	require.Len(t, shelves, 1)

	all, err := repo.List()
	require.NoError(t, err)
	require.NotEmpty(t, all)

	require.NoError(t, repo.Delete(updated))
	deleted, err := repo.Get(id)
	require.NoError(t, err)
	require.Nil(t, deleted)
}

func Test_RealDB_ShelfRepository_UnsetPathAndDomain_AllowMultiple(t *testing.T) {
	repo := TestRepository.ShelfRepository
	userId := realDBTestUser(t)

	id1, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "A"}, UserId: userId})
	require.NoError(t, err)
	id2, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "B"}, UserId: userId})
	require.NoError(t, err)
	require.NotEqual(t, id1, id2, "two shelves with no path/domain must both be creatable")
}

func Test_RealDB_ShelfRepository_DuplicatePath_IsRejected(t *testing.T) {
	repo := TestRepository.ShelfRepository
	userId := realDBTestUser(t)
	path := "taken-" + uuid.NewString()

	_, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "First", Path: path}, UserId: userId})
	require.NoError(t, err)

	_, err = repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "Second", Path: path}, UserId: userId})
	require.Error(t, err, "a second shelf with the same non-empty path must be rejected by the DB's unique constraint")
}

func Test_RealDB_ShelfRepository_DuplicateDomain_IsRejected(t *testing.T) {
	repo := TestRepository.ShelfRepository
	userId := realDBTestUser(t)
	domain := "taken-" + uuid.NewString() + ".example.com"

	_, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "First"}, Domain: domain, UserId: userId})
	require.NoError(t, err)

	_, err = repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "Second"}, Domain: domain, UserId: userId})
	require.Error(t, err, "a second shelf with the same non-empty domain must be rejected by the DB's unique constraint")
}
