package mapper

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MapLinkBaseToLinkPointer(t *testing.T) {
	base := model.LinkBase{
		Title:     "test-title",
		Link:      "https://example.com",
		Icon:      "icon-test",
		Color:     "#ff0000",
		SectionId: "section-uuid-test",
	}

	result := MapLinkBaseToLinkPointer(base)

	require.NotNil(t, result)
	require.Equal(t, base.Title, result.Title)
	require.Equal(t, base.Link, result.Link)
	require.Equal(t, base.Icon, result.Icon)
	require.Equal(t, base.Color, result.Color)
	require.Equal(t, base.SectionId, result.SectionId)
}

func Test_MapLinkToLinkResponse(t *testing.T) {
	link := model.Link{
		Id: "link-uuid-test",
		LinkBase: model.LinkBase{
			Title: "test-title",
		},
	}

	resp := MapLinkToLinkResponse(link)

	require.NotNil(t, resp)
	require.Equal(t, link.Id, resp.Body.Id)
	require.Equal(t, link.Title, resp.Body.Title)
}

func Test_MapLinksToLinkResponseList(t *testing.T) {
	links := []model.Link{
		{
			Id: "link-1",
			LinkBase: model.LinkBase{
				Title: "link-one",
			},
		},
		{
			Id: "link-2",
			LinkBase: model.LinkBase{
				Title: "link-two",
			},
		},
	}

	resp := MapLinksToLinkResponseList(links)

	require.NotNil(t, resp)
	require.Len(t, resp.Body, 2)
	require.Equal(t, "link-1", resp.Body[0].Id)
	require.Equal(t, "link-two", resp.Body[1].Title)
}
