package mapper

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MapShelfBaseToShelfPointer(t *testing.T) {
	base := model.ShelfBase{
		Title:  "test-shelf",
		UserId: "user-uuid-test",
	}

	result := MapShelfBaseToShelfPointer(base)

	require.NotNil(t, result)
	require.Equal(t, base.Title, result.Title)
	require.Equal(t, base.UserId, result.UserId)
}

func Test_MapShelfToShelfResponse(t *testing.T) {
	shelf := model.Shelf{
		PublicShelf: model.PublicShelf{
			Id:    "shelf-uuid-test",
			Title: "test-shelf",
		},
		UserId: "user-uuid-test",
	}

	resp := MapShelfToShelfResponse(shelf)

	require.NotNil(t, resp)
	require.Equal(t, shelf.Id, resp.Body.Id)
	require.Equal(t, shelf.Title, resp.Body.Title)
	require.Equal(t, shelf.UserId, resp.Body.UserId)
}

func Test_MapShelfToPublicShelfResponse(t *testing.T) {
	shelf := model.Shelf{
		PublicShelf: model.PublicShelf{
			Id:          "shelf-uuid-test",
			Title:       "test-shelf",
			Description: "a description",
			Icon:        "i-lucide-book-open",
			Path:        "test-path",
		},
		Domain: "example.com",
		Theme:  "some-theme",
		UserId: "user-uuid-test",
	}

	resp := MapShelfToPublicShelfResponse(shelf)

	require.NotNil(t, resp)
	require.Equal(t, shelf.Id, resp.Body.Id)
	require.Equal(t, shelf.Title, resp.Body.Title)
	require.Equal(t, shelf.Description, resp.Body.Description)
	require.Equal(t, shelf.Icon, resp.Body.Icon)
	require.Equal(t, shelf.Path, resp.Body.Path)
}
