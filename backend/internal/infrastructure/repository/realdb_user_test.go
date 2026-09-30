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
	}, "hashed-password", "user")
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

	_, err := repo.Create(model.UserBase{Email: email, FirstName: "A", LastName: "A"}, "hashed", "user")
	require.NoError(t, err)

	_, err = repo.Create(model.UserBase{Email: email, FirstName: "B", LastName: "B"}, "hashed", "user")
	require.Error(t, err, "a second user with the same email must be rejected by the DB's unique constraint")
}

// An email change waits in pending_email and only replaces the email, with
// the verified flag the caller decides, through ChangeEmail.
func Test_RealDB_UserRepository_PendingEmailAndChangeEmail(t *testing.T) {
	repo := TestRepository.UserRepository
	email := "realdb-pending-" + uuid.NewString() + "@example.com"
	newEmail := "realdb-pending-new-" + uuid.NewString() + "@example.com"

	id, err := repo.Create(model.UserBase{Email: email, FirstName: "Pending", LastName: "Email"}, "hashed", "user")
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

// Regression for the case-sensitive UNIQUE(email) on Postgres: two accounts
// whose emails only differ in case must not both exist.
func Test_RealDB_UserRepository_DuplicateEmail_DifferentCase_IsRejected(t *testing.T) {
	repo := TestRepository.UserRepository
	local := "realdb-case-" + uuid.NewString()

	_, err := repo.Create(model.UserBase{Email: local + "@example.com", FirstName: "A", LastName: "A"}, "hashed", "user")
	require.NoError(t, err)

	_, err = repo.Create(model.UserBase{Email: strings.ToUpper(local) + "@EXAMPLE.COM", FirstName: "B", LastName: "B"}, "hashed", "user")
	require.Error(t, err)
}

func Test_RealDB_UserRepository_AuthMethods(t *testing.T) {
	repo := TestRepository.UserRepository
	email := "realdb-auth-" + uuid.NewString() + "@example.com"

	id, err := repo.Create(model.UserBase{Email: email, FirstName: "Auth", LastName: "Flow"}, "hashed", "user")
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
	externalId, err := repo.CreateExternal(externalEmail, "external-"+uuid.NewString()[:8], "External", "User", "OIDC", externalProviderId)
	require.NoError(t, err)
	require.NotEmpty(t, externalId)

	externalUser, err := repo.FindByProviderId(externalProviderId)
	require.NoError(t, err)
	require.NotNil(t, externalUser)
	require.Equal(t, externalEmail, externalUser.Email)
	require.Equal(t, "OIDC", externalUser.Provider)
}
