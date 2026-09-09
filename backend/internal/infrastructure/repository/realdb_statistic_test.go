//go:build realdb

package repository

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RealDB_StatisticRepository_Counts(t *testing.T) {
	userId := realDBTestUser(t)

	shelfId, err := TestRepository.ShelfRepository.Create(&model.Shelf{
		PublicShelf: model.PublicShelf{Title: "Stats shelf"},
		UserId:      userId,
	})
	require.NoError(t, err)

	sectionId, err := TestRepository.SectionRepository.Create(&model.Section{
		SectionBase: model.SectionBase{Title: "Stats section", ShelfId: shelfId},
	})
	require.NoError(t, err)

	_, err = TestRepository.LinkRepository.Create(&model.Link{LinkBase: model.LinkBase{
		Title: "Stats link", Link: "https://example.com", SectionId: sectionId,
	}})
	require.NoError(t, err)

	shelfCount, err := TestRepository.StatisticRepository.GetShelfAmount(userId)
	require.NoError(t, err)
	require.NotNil(t, shelfCount)
	require.Equal(t, 1, *shelfCount)

	sectionCount, err := TestRepository.StatisticRepository.GetSectionAmount(userId)
	require.NoError(t, err)
	require.NotNil(t, sectionCount)
	require.Equal(t, 1, *sectionCount)

	linkCount, err := TestRepository.StatisticRepository.GetLinkAmount(userId)
	require.NoError(t, err)
	require.NotNil(t, linkCount)
	require.Equal(t, 1, *linkCount)
}
