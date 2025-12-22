package mapper

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MapSectionBaseToSectionPointer(t *testing.T) {
	base := model.SectionBase{
		Title:   "test-section",
		ShelfId: "shelf-uuid-test",
	}

	result := MapSectionBaseToSectionPointer(base)

	require.NotNil(t, result)
	require.Equal(t, base.Title, result.Title)
	require.Equal(t, base.ShelfId, result.ShelfId)
}

func Test_MapSectionToSectionResponse(t *testing.T) {
	section := model.Section{
		Id: "section-uuid-test",
		SectionBase: model.SectionBase{
			Title: "test-section",
		},
	}

	resp := MapSectionToSectionResponse(section)

	require.NotNil(t, resp)
	require.Equal(t, section.Id, resp.Body.Id)
	require.Equal(t, section.Title, resp.Body.Title)
}

func Test_MapSectionsToSectionResponseList(t *testing.T) {
	sections := []model.Section{
		{
			Id: "section-1",
			SectionBase: model.SectionBase{
				Title: "section-one",
			},
		},
		{
			Id: "section-2",
			SectionBase: model.SectionBase{
				Title: "section-two",
			},
		},
	}

	resp := MapSectionsToSectionResponseList(sections)

	require.NotNil(t, resp)
	require.Len(t, resp.Body, 2)
	require.Equal(t, "section-1", resp.Body[0].Id)
	require.Equal(t, "section-two", resp.Body[1].Title)
}
