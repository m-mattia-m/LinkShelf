//go:build realdb

package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_RealDB_RefreshTokenRepository_CRUD(t *testing.T) {
	repo := TestRepository.RefreshTokenRepository
	userId := realDBTestUser(t)
	hash := "hash-" + uuid.NewString()

	require.NoError(t, repo.Create(userId, hash, time.Now().Add(time.Hour)))

	fetched, err := repo.GetByHash(hash)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, userId, fetched.UserId)
	require.WithinDuration(t, time.Now().Add(time.Hour), fetched.ExpiresAt, 5*time.Second)

	missing, err := repo.GetByHash("does-not-exist-" + uuid.NewString())
	require.NoError(t, err)
	require.Nil(t, missing)

	require.NoError(t, repo.DeleteByHash(hash))
	afterDelete, err := repo.GetByHash(hash)
	require.NoError(t, err)
	require.Nil(t, afterDelete)
}

func Test_RealDB_RefreshTokenRepository_DeleteByUserId(t *testing.T) {
	repo := TestRepository.RefreshTokenRepository
	userId := realDBTestUser(t)

	hash1 := "hash-" + uuid.NewString()
	hash2 := "hash-" + uuid.NewString()
	require.NoError(t, repo.Create(userId, hash1, time.Now().Add(time.Hour)))
	require.NoError(t, repo.Create(userId, hash2, time.Now().Add(time.Hour)))

	require.NoError(t, repo.DeleteByUserId(userId))

	t1, err := repo.GetByHash(hash1)
	require.NoError(t, err)
	require.Nil(t, t1)
	t2, err := repo.GetByHash(hash2)
	require.NoError(t, err)
	require.Nil(t, t2)
}

func Test_RealDB_OidcStateRepository_CRUD(t *testing.T) {
	repo := TestRepository.OidcStateRepository
	state := "state-" + uuid.NewString()

	require.NoError(t, repo.Create(state, "verifier-value", time.Now().Add(10*time.Minute)))

	fetched, err := repo.GetByState(state)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, "verifier-value", fetched.CodeVerifier)

	missing, err := repo.GetByState("does-not-exist-" + uuid.NewString())
	require.NoError(t, err)
	require.Nil(t, missing)

	require.NoError(t, repo.DeleteByState(state))
	afterDelete, err := repo.GetByState(state)
	require.NoError(t, err)
	require.Nil(t, afterDelete)
}
