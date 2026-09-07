//go:build realdb

package repository

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func realDBTestShelf(t *testing.T) string {
	t.Helper()
	userId := realDBTestUser(t)
	id, err := TestRepository.ShelfRepository.Create(&model.Shelf{
		PublicShelf: model.PublicShelf{Title: "Section/Link fixture shelf"},
		UserId:      userId,
	})
	require.NoError(t, err)
	return id
}

func Test_RealDB_SectionRepository_CRUD(t *testing.T) {
	repo := TestRepository.SectionRepository
	shelfId := realDBTestShelf(t)

	id, err := repo.Create(&model.Section{SectionBase: model.SectionBase{Title: "Section 1", ShelfId: shelfId}})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	fetched, err := repo.Get(id)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, "Section 1", fetched.Title)

	fetched.Title = "Renamed Section"
	require.NoError(t, repo.Update(fetched))

	updated, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, "Renamed Section", updated.Title)

	sections, err := repo.ListByShelfId(shelfId)
	require.NoError(t, err)
	require.Len(t, sections, 1)

	require.NoError(t, repo.Delete(updated))
	deleted, err := repo.Get(id)
	require.NoError(t, err)
	require.Nil(t, deleted)
}

func Test_RealDB_LinkRepository_CRUD(t *testing.T) {
	shelfId := realDBTestShelf(t)
	sectionId, err := TestRepository.SectionRepository.Create(&model.Section{
		SectionBase: model.SectionBase{Title: "Link fixture section", ShelfId: shelfId},
	})
	require.NoError(t, err)

	repo := TestRepository.LinkRepository

	id, err := repo.Create(&model.Link{LinkBase: model.LinkBase{
		Title:     "My Link",
		Link:      "https://example.com",
		Icon:      "link",
		Color:     "#123456",
		SectionId: sectionId,
	}})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	fetched, err := repo.Get(id)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, "https://example.com", fetched.Link)

	fetched.Title = "Renamed Link"
	fetched.Link = "https://example.org"
	require.NoError(t, repo.Update(fetched))

	updated, err := repo.Get(id)
	require.NoError(t, err)
	require.Equal(t, "Renamed Link", updated.Title)
	require.Equal(t, "https://example.org", updated.Link)

	links, err := repo.ListByShelfId(shelfId)
	require.NoError(t, err)
	require.Len(t, links, 1)

	require.NoError(t, repo.Delete(updated))
	deleted, err := repo.Get(id)
	require.NoError(t, err)
	require.Nil(t, deleted)
}

func Test_RealDB_SectionDelete_CascadesToLinks(t *testing.T) {
	shelfId := realDBTestShelf(t)
	sectionId, err := TestRepository.SectionRepository.Create(&model.Section{
		SectionBase: model.SectionBase{Title: "Cascade section", ShelfId: shelfId},
	})
	require.NoError(t, err)

	linkId, err := TestRepository.LinkRepository.Create(&model.Link{LinkBase: model.LinkBase{
		Title: "Cascade link", Link: "https://example.com", SectionId: sectionId,
	}})
	require.NoError(t, err)

	require.NoError(t, TestRepository.SectionRepository.Delete(&model.Section{Id: sectionId}))

	link, err := TestRepository.LinkRepository.Get(linkId)
	require.NoError(t, err)
	require.Nil(t, link, "deleting a section must cascade-delete its links on both engines")
}
