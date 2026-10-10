//go:build realdb

package repository

import (
	"backend/internal/infrastructure/api/model"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_RealDB_UserRepository_CRUD(t *testing.T) {
	repo := TestRepository.UserRepository
	email := "realdb-" + uuid.NewString() + "@example.com"

	id, err := repo.Create(model.UserBase{
		Email:     email,
		FirstName: "Real",
		LastName:  "DB",
	}, "hashed-password", "user", nil)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	fetched, err := repo.Get(id)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, email, fetched.Email)
	require.Equal(t, "user", fetched.Role)

	password, err := repo.GetPassword(id)
	require.NoError(t, err)
	require.Equal(t, "hashed-password", password)

	users, err := repo.List()
	require.NoError(t, err)
	require.NotEmpty(t, users)

	fetched.FirstName = "Updated"
	fetched.Role = "admin"
	require.NoError(t, repo.Update(fetched))

	updated, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, "Updated", updated.FirstName)
	require.Equal(t, "admin", updated.Role)

	require.NoError(t, repo.PatchPassword(id, "new-hashed-password"))
	password, err = repo.GetPassword(id)
	require.NoError(t, err)
	require.Equal(t, "new-hashed-password", password)

	require.NoError(t, repo.Delete(updated))
	deleted, err := repo.Get(id)
	require.NoError(t, err)
	require.Nil(t, deleted)
}

func Test_RealDB_UserRepository_DuplicateEmail_IsRejected(t *testing.T) {
	repo := TestRepository.UserRepository
	email := "realdb-dup-" + uuid.NewString() + "@example.com"

	_, err := repo.Create(model.UserBase{Email: email, FirstName: "A", LastName: "A"}, "hashed", "user", nil)
	require.NoError(t, err)

	_, err = repo.Create(model.UserBase{Email: email, FirstName: "B", LastName: "B"}, "hashed", "user", nil)
	require.Error(t, err, "a second user with the same email must be rejected by the DB's unique constraint")
}

// A pending email only replaces the email through ChangeEmail.
func Test_RealDB_UserRepository_PendingEmailAndChangeEmail(t *testing.T) {
	repo := TestRepository.UserRepository
	email := "realdb-pending-" + uuid.NewString() + "@example.com"
	newEmail := "realdb-pending-new-" + uuid.NewString() + "@example.com"

	id, err := repo.Create(model.UserBase{Email: email, FirstName: "Pending", LastName: "Email"}, "hashed", "user", nil)
	require.NoError(t, err)
	require.NoError(t, repo.MarkVerified(id))

	require.NoError(t, repo.SetPendingEmail(id, newEmail))
	pending, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, email, pending.Email)
	require.Equal(t, newEmail, pending.PendingEmail)
	require.True(t, pending.EmailVerified)

	require.NoError(t, repo.ChangeEmail(id, newEmail, false))
	changed, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, newEmail, changed.Email)
	require.Empty(t, changed.PendingEmail)
	require.False(t, changed.EmailVerified)

	require.NoError(t, repo.ChangeEmail(id, email, true))
	confirmed, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, email, confirmed.Email)
	require.True(t, confirmed.EmailVerified)

	require.NoError(t, repo.SetPendingEmail(id, newEmail))
	require.NoError(t, repo.SetPendingEmail(id, ""))
	cleared, err := repo.Get(id)
	require.NoError(t, err)
	require.Empty(t, cleared.PendingEmail)
}

// Regression: emails differing only in case can't both exist.
func Test_RealDB_UserRepository_DuplicateEmail_DifferentCase_IsRejected(t *testing.T) {
	repo := TestRepository.UserRepository
	local := "realdb-case-" + uuid.NewString()

	_, err := repo.Create(model.UserBase{Email: local + "@example.com", FirstName: "A", LastName: "A"}, "hashed", "user", nil)
	require.NoError(t, err)

	_, err = repo.Create(model.UserBase{Email: strings.ToUpper(local) + "@EXAMPLE.COM", FirstName: "B", LastName: "B"}, "hashed", "user", nil)
	require.Error(t, err)
}

func Test_RealDB_UserRepository_AuthMethods(t *testing.T) {
	repo := TestRepository.UserRepository
	email := "realdb-auth-" + uuid.NewString() + "@example.com"

	id, err := repo.Create(model.UserBase{Email: email, FirstName: "Auth", LastName: "Flow"}, "hashed", "user", nil)
	require.NoError(t, err)

	byEmail, err := repo.FindByEmail(email)
	require.NoError(t, err)
	require.NotNil(t, byEmail)
	require.Equal(t, id, byEmail.Id)
	require.Equal(t, "LOCAL", byEmail.Provider)

	notFound, err := repo.FindByEmail("does-not-exist-" + uuid.NewString() + "@example.com")
	require.NoError(t, err)
	require.Nil(t, notFound)

	providerId := "oidc-sub-" + uuid.NewString()
	require.NoError(t, repo.LinkProvider(id, "OIDC", providerId))

	byProvider, err := repo.FindByProviderId(providerId)
	require.NoError(t, err)
	require.NotNil(t, byProvider)
	require.Equal(t, id, byProvider.Id)
	require.Equal(t, "OIDC", byProvider.Provider)

	externalEmail := "realdb-external-" + uuid.NewString() + "@example.com"
	externalProviderId := "oidc-sub-" + uuid.NewString()
	externalId, err := repo.CreateExternal(externalEmail, "external-"+uuid.NewString()[:8], "External", "User", "OIDC", externalProviderId, nil)
	require.NoError(t, err)
	require.NotEmpty(t, externalId)

	externalUser, err := repo.FindByProviderId(externalProviderId)
	require.NoError(t, err)
	require.NotNil(t, externalUser)
	require.Equal(t, externalEmail, externalUser.Email)
	require.Equal(t, "OIDC", externalUser.Provider)
}

func Test_RealDB_UserRepository_MaxShelves(t *testing.T) {
	repo := TestRepository.UserRepository
	limit := 2

	id, err := repo.Create(model.UserBase{
		Email:     "realdb-" + uuid.NewString() + "@example.com",
		FirstName: "Lim",
		LastName:  "It",
	}, "hashed", "user", &limit)
	require.NoError(t, err)

	user, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, 2, *user.MaxShelves)

	// Unchanged value still reports the user as found (MySQL affects 0 rows).
	found, err := repo.SetMaxShelves(id, &limit)
	require.NoError(t, err)
	require.True(t, found)

	found, err = repo.SetMaxShelves(id, nil)
	require.NoError(t, err)
	require.True(t, found)
	user, err = repo.Get(id)
	require.NoError(t, err)
	require.Nil(t, user.MaxShelves)

	zero := 0
	_, err = repo.SetMaxShelves(id, &zero)
	require.NoError(t, err)
	user, _ = repo.Get(id)
	require.Equal(t, 0, *user.MaxShelves)

	found, err = repo.SetMaxShelves(uuid.NewString(), nil)
	require.NoError(t, err)
	require.False(t, found)
}

func Test_RealDB_ShelfRepository_CountByUserId(t *testing.T) {
	userId := realDBTestUser(t)
	repo := TestRepository.ShelfRepository

	count, err := repo.CountByUserId(userId)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	_, err = repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "A"}, UserId: userId})
	require.NoError(t, err)
	count, err = repo.CountByUserId(userId)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
