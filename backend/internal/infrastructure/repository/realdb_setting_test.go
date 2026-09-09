//go:build realdb

package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_RealDB_SettingRepository_UpsertAndGetByKey(t *testing.T) {
	repo := TestRepository.SettingRepository
	key := "realdb-setting-" + uuid.NewString()

	require.NoError(t, repo.Upsert(key, "en", "first value"))

	setting, err := repo.GetByKey(key)
	require.NoError(t, err)
	require.NotNil(t, setting)
	require.Equal(t, "first value", setting.Value)

	// Upsert must update on a repeated key/language, not insert a duplicate.
	require.NoError(t, repo.Upsert(key, "en", "second value"))

	updated, err := repo.GetByKey(key)
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, "second value", updated.Value)

	missing, err := repo.GetByKey("realdb-setting-does-not-exist-" + uuid.NewString())
	require.NoError(t, err)
	require.Nil(t, missing)

	all, err := repo.List()
	require.NoError(t, err)
	require.NotEmpty(t, all)
}
