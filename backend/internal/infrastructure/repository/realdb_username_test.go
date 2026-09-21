//go:build realdb

package repository

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func realDBUsername() string {
	return "realdb-" + uuid.NewString()[:12]
}

func realDBUserWithUsername(t *testing.T, username string) string {
	t.Helper()
	id, err := TestRepository.UserRepository.Create(model.UserBase{
		Email:     "realdb-" + uuid.NewString() + "@example.com",
		Username:  username,
		FirstName: "Real",
		LastName:  "DB",
	}, "hashed", "user")
	require.NoError(t, err)
	return id
}

func Test_RealDB_UserRepository_Username(t *testing.T) {
	repo := TestRepository.UserRepository
	username := realDBUsername()
	id := realDBUserWithUsername(t, username)

	fetched, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, username, fetched.Username)

	all, err := repo.List()
	require.NoError(t, err)
	var listed *model.User
	for i := range all {
		if all[i].Id == id {
			listed = &all[i]
		}
	}
	require.NotNil(t, listed)
	require.Equal(t, username, listed.Username)

	record, err := repo.FindByEmail(fetched.Email)
	require.NoError(t, err)
	require.Equal(t, username, record.Username)

	taken, err := repo.UsernameTaken(username, "")
	require.NoError(t, err)
	require.True(t, taken)

	// Renaming to your own current name is not a clash.
	taken, err = repo.UsernameTaken(username, id)
	require.NoError(t, err)
	require.False(t, taken)

	taken, err = repo.UsernameTaken(realDBUsername(), "")
	require.NoError(t, err)
	require.False(t, taken)
}

func Test_RealDB_UserRepository_UsernameIsUnique(t *testing.T) {
	username := realDBUsername()
	realDBUserWithUsername(t, username)

	_, err := TestRepository.UserRepository.Create(model.UserBase{
		Email:     "realdb-" + uuid.NewString() + "@example.com",
		Username:  username,
		FirstName: "Other",
		LastName:  "User",
	}, "hashed", "user")

	require.Error(t, err, "the same username cannot belong to two users")
}

func Test_RealDB_UserRepository_UsersWithoutUsernameDoNotCollide(t *testing.T) {
	repo := TestRepository.UserRepository

	// Accounts from before usernames existed all have NULL, which a UNIQUE
	// constraint must let through on both engines.
	first := realDBUserWithUsername(t, "")
	second := realDBUserWithUsername(t, "")

	missing, err := repo.ListWithoutUsername()
	require.NoError(t, err)
	ids := map[string]string{}
	for _, user := range missing {
		ids[user.Id] = user.Email
	}
	require.Contains(t, ids, first)
	require.Contains(t, ids, second)
	require.NotEmpty(t, ids[first], "the email comes back so a username can be derived from it")

	name := realDBUsername()
	require.NoError(t, repo.SetUsername(first, name))

	fetched, err := repo.Get(first)
	require.NoError(t, err)
	require.Equal(t, name, fetched.Username)

	missing, err = repo.ListWithoutUsername()
	require.NoError(t, err)
	for _, user := range missing {
		require.NotEqual(t, first, user.Id, "a user that got a username is no longer listed")
	}
}

func Test_RealDB_UserRepository_UpdateChangesTheUsername(t *testing.T) {
	repo := TestRepository.UserRepository
	id := realDBUserWithUsername(t, realDBUsername())

	fetched, err := repo.Get(id)
	require.NoError(t, err)

	renamed := realDBUsername()
	fetched.Username = renamed
	require.NoError(t, repo.Update(fetched))

	after, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, renamed, after.Username)
}

func Test_RealDB_UserRepository_CreateExternalStoresTheUsername(t *testing.T) {
	repo := TestRepository.UserRepository
	username := realDBUsername()
	providerId := "oidc-" + uuid.NewString()

	id, err := repo.CreateExternal("realdb-"+uuid.NewString()+"@example.com", username, "Ext", "User", "OIDC", providerId)
	require.NoError(t, err)

	fetched, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, username, fetched.Username)

	record, err := repo.FindByProviderId(providerId)
	require.NoError(t, err)
	require.Equal(t, username, record.Username)
}

func Test_RealDB_ShelfRepository_TwoOwnersMayShareAPath(t *testing.T) {
	repo := TestRepository.ShelfRepository
	aliceName, bobName := realDBUsername(), realDBUsername()
	alice := realDBUserWithUsername(t, aliceName)
	bob := realDBUserWithUsername(t, bobName)
	path := "shared-" + uuid.NewString()[:8]

	aliceShelf, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "Alice", Path: path}, UserId: alice})
	require.NoError(t, err)
	bobShelf, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "Bob", Path: path}, UserId: bob})
	require.NoError(t, err, "the per-owner constraint lets two users use the same path")

	// ... but not the same owner twice.
	_, err = repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "Alice again", Path: path}, UserId: alice})
	require.Error(t, err)

	// Shelves without a path never collide.
	_, err = repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "No path 1"}, UserId: alice})
	require.NoError(t, err)
	_, err = repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "No path 2"}, UserId: alice})
	require.NoError(t, err)

	// The owner's username comes back on every read.
	fetched, err := repo.Get(aliceShelf)
	require.NoError(t, err)
	require.Equal(t, aliceName, fetched.Username)

	mine, err := repo.ListByUserId(bob)
	require.NoError(t, err)
	require.Len(t, mine, 1)
	require.Equal(t, bobName, mine[0].Username)

	all, err := repo.List()
	require.NoError(t, err)
	found := map[string]string{}
	for _, shelf := range all {
		found[shelf.Id] = shelf.Username
	}
	require.Equal(t, aliceName, found[aliceShelf])
	require.Equal(t, bobName, found[bobShelf])

	// Lookup by username and path, case-insensitive, scoped to the owner.
	byUser, err := repo.GetByUsernameAndPath(aliceName, path)
	require.NoError(t, err)
	require.NotNil(t, byUser)
	require.Equal(t, aliceShelf, byUser.Id)

	byUser, err = repo.GetByUsernameAndPath(bobName, "SHARED-"+path[len("shared-"):])
	require.NoError(t, err)
	require.NotNil(t, byUser)
	require.Equal(t, bobShelf, byUser.Id)

	missing, err := repo.GetByUsernameAndPath("nobody-"+uuid.NewString()[:8], path)
	require.NoError(t, err)
	require.Nil(t, missing)
}

func Test_RealDB_ShelfRepository_PathInUse(t *testing.T) {
	repo := TestRepository.ShelfRepository
	alice := realDBUserWithUsername(t, realDBUsername())
	bob := realDBUserWithUsername(t, realDBUsername())
	path := "inuse-" + uuid.NewString()[:8]

	aliceShelf, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "A", Path: path}, UserId: alice})
	require.NoError(t, err)

	taken, err := repo.PathInUse(path, "")
	require.NoError(t, err)
	require.True(t, taken)

	taken, err = repo.PathInUse("INUSE-"+path[len("inuse-"):], "")
	require.NoError(t, err)
	require.True(t, taken, "compared case-insensitively, like the public lookup")

	taken, err = repo.PathInUse(path, aliceShelf)
	require.NoError(t, err)
	require.False(t, taken, "a shelf does not clash with itself")

	taken, err = repo.PathInUse("free-"+uuid.NewString()[:8], "")
	require.NoError(t, err)
	require.False(t, taken)

	taken, err = repo.PathInUseByUser(alice, path, "")
	require.NoError(t, err)
	require.True(t, taken)

	taken, err = repo.PathInUseByUser(alice, path, aliceShelf)
	require.NoError(t, err)
	require.False(t, taken)

	taken, err = repo.PathInUseByUser(bob, path, "")
	require.NoError(t, err)
	require.False(t, taken, "another owner's shelf does not count")
}

func Test_RealDB_ShelfRepository_ListPathCollisions(t *testing.T) {
	repo := TestRepository.ShelfRepository
	aliceName, bobName := realDBUsername(), realDBUsername()
	alice := realDBUserWithUsername(t, aliceName)
	bob := realDBUserWithUsername(t, bobName)
	suffix := uuid.NewString()[:8]

	// A collision differing only in case, as the lookup treats it as one path.
	aliceShelf, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "A", Path: "clash-" + suffix}, UserId: alice})
	require.NoError(t, err)
	bobShelf, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "B", Path: "CLASH-" + suffix}, UserId: bob})
	require.NoError(t, err)
	unique, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "U", Path: "unique-" + suffix}, UserId: alice})
	require.NoError(t, err)

	collisions, err := repo.ListPathCollisions()
	require.NoError(t, err)

	byShelf := map[string]PathCollision{}
	for _, c := range collisions {
		byShelf[c.ShelfId] = c
	}
	require.Contains(t, byShelf, aliceShelf)
	require.Contains(t, byShelf, bobShelf)
	require.NotContains(t, byShelf, unique)
	require.Equal(t, aliceName, byShelf[aliceShelf].Username)
	require.Equal(t, alice, byShelf[aliceShelf].UserId)
	require.Equal(t, "CLASH-"+suffix, byShelf[bobShelf].Path)
	require.Equal(t, bobName, byShelf[bobShelf].Username)
}

func Test_RealDB_ShelfRepository_StoresTheCreationModeAndUpdateLeavesItAlone(t *testing.T) {
	repo := TestRepository.ShelfRepository
	owner := realDBUserWithUsername(t, realDBUsername())

	off, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "Off"}, UserId: owner})
	require.NoError(t, err)
	on, err := repo.Create(&model.Shelf{PublicShelf: model.PublicShelf{Title: "On"}, UserId: owner, CreatedWithUserBasedPaths: true})
	require.NoError(t, err)

	offShelf, err := repo.Get(off)
	require.NoError(t, err)
	require.False(t, offShelf.CreatedWithUserBasedPaths, "the column defaults to off for shelves that predate the setting")
	onShelf, err := repo.Get(on)
	require.NoError(t, err)
	require.True(t, onShelf.CreatedWithUserBasedPaths)

	// An update - even one carrying the opposite value - never rewrites it.
	offShelf.Title = "Off, edited"
	offShelf.CreatedWithUserBasedPaths = true
	require.NoError(t, repo.Update(offShelf))
	onShelf.Title = "On, edited"
	onShelf.CreatedWithUserBasedPaths = false
	require.NoError(t, repo.Update(onShelf))

	offAfter, err := repo.Get(off)
	require.NoError(t, err)
	require.Equal(t, "Off, edited", offAfter.Title)
	require.False(t, offAfter.CreatedWithUserBasedPaths)
	onAfter, err := repo.Get(on)
	require.NoError(t, err)
	require.Equal(t, "On, edited", onAfter.Title)
	require.True(t, onAfter.CreatedWithUserBasedPaths)

	// The list and username lookups carry it too.
	all, err := repo.List()
	require.NoError(t, err)
	byId := map[string]bool{}
	for _, shelf := range all {
		byId[shelf.Id] = shelf.CreatedWithUserBasedPaths
	}
	require.False(t, byId[off])
	require.True(t, byId[on])
}
