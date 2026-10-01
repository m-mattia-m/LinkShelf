package domain

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Unit_Link_List_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	links := []model.Link{
		{Id: "link-1"},
		{Id: "link-2"},
	}

	svc.LinkRepository.
		EXPECT().
		ListByShelfId("shelf-uuid-test").
		Return(links, nil)

	result, err := svc.Service.LinkService.List("shelf-uuid-test")

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "link-1", result[0].Id)
}

func Test_Unit_Link_List_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		ListByShelfId("shelf-uuid-test").
		Return(nil, errors.New("an error occurred"))

	links, err := svc.Service.LinkService.List("shelf-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, links)
}

func Test_Unit_Link_Get_Success_Trims_Color(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("link-uuid-test").
		Return(&model.Link{
			Id: "link-uuid-test",
			LinkBase: model.LinkBase{
				Title:     "link-title-test",
				Link:      "https://example.com",
				Icon:      "my-base64-encoded-icon",
				Color:     "#FFFFFF",
				SectionId: "c5f3738e-668e-409e-bccd-c5c1b31de0da",
			},
		}, nil)

	link, err := svc.Service.LinkService.Get("link-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, link)
	require.Equal(t, "#FFFFFF", link.Color)
}

func Test_Unit_Link_Get_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("link-uuid-test").
		Return(nil, errors.New("an error occurred"))

	link, err := svc.Service.LinkService.Get("link-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, link)
}

func Test_Unit_Link_Create_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkRequest := &model.Link{
		LinkBase: model.LinkBase{
			Title:     "link-title-test",
			Link:      "https://example.com",
			SectionId: "section-uuid-test",
		},
	}

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.LinkRepository.
		EXPECT().
		Create(linkRequest).
		Return("link-uuid-test", nil)

	svc.LinkRepository.
		EXPECT().
		Get("link-uuid-test").
		Return(&model.Link{
			Id: "link-uuid-test",
			LinkBase: model.LinkBase{
				Title:     "link-title-test",
				Link:      "https://example.com",
				SectionId: "section-uuid-test",
			},
		}, nil)

	link, err := svc.Service.LinkService.Create("user-uuid-test", false, linkRequest)

	require.NoError(t, err)
	require.NotNil(t, link)
	require.Equal(t, "link-uuid-test", link.Id)
	require.Equal(t, "link-title-test", link.Title)
}

func Test_Unit_Link_Create_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkRequest := &model.Link{LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	link, err := svc.Service.LinkService.Create("someone-else-uuid-test", false, linkRequest)

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, link)
}

func Test_Unit_Link_Create_NotFound_Section(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkRequest := &model.Link{LinkBase: model.LinkBase{SectionId: "missing-section"}}

	svc.SectionRepository.
		EXPECT().
		Get("missing-section").
		Return(nil, nil)

	link, err := svc.Service.LinkService.Create("user-uuid-test", false, linkRequest)

	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, link)
}

func Test_Unit_Link_Create_InvalidURL(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkRequest := &model.Link{
		LinkBase: model.LinkBase{
			Link:      "not a url",
			SectionId: "section-uuid-test",
		},
	}

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	link, err := svc.Service.LinkService.Create("user-uuid-test", false, linkRequest)

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, link)
}

func Test_Unit_Link_Update_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkId := "link-uuid-test"

	updateRequest := &model.Link{
		LinkBase: model.LinkBase{
			Title:     "link-title-test-updated",
			Link:      "https://updated.example.com",
			SectionId: "section-uuid-test",
		},
	}

	svc.LinkRepository.
		EXPECT().
		Get(linkId).
		Return(&model.Link{Id: linkId, LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.LinkRepository.
		EXPECT().
		Update(&model.Link{
			Id:       linkId,
			LinkBase: updateRequest.LinkBase,
		}).
		Return(nil)

	svc.LinkRepository.
		EXPECT().
		Get(linkId).
		Return(&model.Link{
			Id: linkId,
			LinkBase: model.LinkBase{
				Title:     "link-title-test-updated",
				Link:      "https://updated.example.com",
				SectionId: "section-uuid-test",
			},
		}, nil)

	link, err := svc.Service.LinkService.Update(linkId, "user-uuid-test", false, updateRequest)

	require.NoError(t, err)
	require.NotNil(t, link)
	require.Equal(t, "link-title-test-updated", link.Title)
}

func Test_Unit_Link_Update_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkId := "link-uuid-test"

	svc.LinkRepository.
		EXPECT().
		Get(linkId).
		Return(&model.Link{Id: linkId, LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	link, err := svc.Service.LinkService.Update(linkId, "someone-else-uuid-test", false, &model.Link{})

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, link)
}

func Test_Unit_Link_Update_NotFound_Section(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkId := "link-uuid-test"

	svc.LinkRepository.
		EXPECT().
		Get(linkId).
		Return(&model.Link{Id: linkId, LinkBase: model.LinkBase{SectionId: "missing-section"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("missing-section").
		Return(nil, nil)

	link, err := svc.Service.LinkService.Update(linkId, "user-uuid-test", false, &model.Link{})

	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, link)
}

func Test_Unit_Link_Update_InvalidURL(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkId := "link-uuid-test"

	updateRequest := &model.Link{
		LinkBase: model.LinkBase{
			Link:      "ftp://example.com",
			SectionId: "section-uuid-test",
		},
	}

	svc.LinkRepository.
		EXPECT().
		Get(linkId).
		Return(&model.Link{Id: linkId, LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	link, err := svc.Service.LinkService.Update(linkId, "user-uuid-test", false, updateRequest)

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, link)
}

func Test_Unit_Link_Delete_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("link-uuid-test").
		Return(&model.Link{Id: "link-uuid-test", LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.LinkRepository.
		EXPECT().
		Delete(&model.Link{Id: "link-uuid-test"}).
		Return(nil)

	err := svc.Service.LinkService.Delete("link-uuid-test", "user-uuid-test", false)

	require.NoError(t, err)
}

func Test_Unit_NormalizeLinkURL(t *testing.T) {
	valid := map[string]string{
		"example.com":                "https://example.com",
		"https://example.com":        "https://example.com",
		"http://example.com":         "http://example.com",
		"sub.example.co.uk/path?x=1": "https://sub.example.co.uk/path?x=1",
		"example.com:8080/path":      "https://example.com:8080/path",
	}
	for value, want := range valid {
		t.Run(value, func(t *testing.T) {
			got, err := normalizeLinkURL(value)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}

	invalid := []string{
		"ftp://example.com",
		"not a url",
		"example",
		"",
	}
	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			_, err := normalizeLinkURL(value)
			require.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

// Regression: javascript: URLs disguised as userinfo must be rejected.
func Test_Unit_NormalizeLinkURL_Rejects_ScriptPayloads(t *testing.T) {
	payloads := []string{
		"javascript:alert(document.domain)%2F%2F@example.com",
		"JavaScript:alert(1)%2F%2F@example.com",
		"data:text/html,<script>alert(1)</script>%2F%2F@example.com",
		"javascript://example.com/%0Aalert(1)",
		"vbscript:msgbox(1)%2F%2F@example.com",
		"https://user:pass@example.com",
		"user@example.com",
	}
	for _, value := range payloads {
		t.Run(value, func(t *testing.T) {
			got, err := normalizeLinkURL(value)
			require.ErrorIs(t, err, ErrInvalidInput)
			require.Empty(t, got)
		})
	}
}

func Test_Unit_Link_Create_Stores_NormalizedURL(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkRequest := &model.Link{LinkBase: model.LinkBase{Link: "example.com/path", SectionId: "section-uuid-test"}}

	svc.SectionRepository.EXPECT().Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)
	svc.ShelfRepository.EXPECT().Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)
	svc.LinkRepository.EXPECT().
		Create(&model.Link{LinkBase: model.LinkBase{Link: "https://example.com/path", SectionId: "section-uuid-test"}}).
		Return("link-uuid-test", nil)
	svc.LinkRepository.EXPECT().Get("link-uuid-test").
		Return(&model.Link{Id: "link-uuid-test", LinkBase: model.LinkBase{Link: "https://example.com/path"}}, nil)

	link, err := svc.Service.LinkService.Create("user-uuid-test", false, linkRequest)

	require.NoError(t, err)
	require.Equal(t, "https://example.com/path", link.Link)
}

func Test_Unit_Link_Create_Rejects_JavascriptPayload(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkRequest := &model.Link{LinkBase: model.LinkBase{Link: "javascript:alert(document.domain)%2F%2F@example.com", SectionId: "section-uuid-test"}}

	svc.SectionRepository.EXPECT().Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)
	svc.ShelfRepository.EXPECT().Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	link, err := svc.Service.LinkService.Create("user-uuid-test", false, linkRequest)

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, link)
}

func Test_Unit_Link_Update_Stores_NormalizedURL(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkId := "link-uuid-test"

	svc.LinkRepository.EXPECT().Get(linkId).
		Return(&model.Link{Id: linkId, LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)
	svc.SectionRepository.EXPECT().Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)
	svc.ShelfRepository.EXPECT().Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)
	svc.LinkRepository.EXPECT().
		Update(&model.Link{Id: linkId, LinkBase: model.LinkBase{Link: "https://example.com", SectionId: "section-uuid-test"}}).
		Return(nil)
	svc.LinkRepository.EXPECT().Get(linkId).
		Return(&model.Link{Id: linkId, LinkBase: model.LinkBase{Link: "https://example.com"}}, nil)

	link, err := svc.Service.LinkService.Update(linkId, "user-uuid-test", false,
		&model.Link{LinkBase: model.LinkBase{Link: "example.com", SectionId: "section-uuid-test"}})

	require.NoError(t, err)
	require.Equal(t, "https://example.com", link.Link)
}

func Test_Unit_Link_Delete_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("link-uuid-test").
		Return(&model.Link{Id: "link-uuid-test", LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	err := svc.Service.LinkService.Delete("link-uuid-test", "someone-else-uuid-test", false)

	require.ErrorIs(t, err, ErrForbidden)
}

func Test_Unit_Link_UpdateOrder_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("link-1").
		Return(&model.Link{Id: "link-1", LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.LinkRepository.
		EXPECT().
		UpdateOrder("link-1", 0).
		Return(nil)

	failures := svc.Service.LinkService.UpdateOrder("user-uuid-test", false, []model.LinkOrderItem{
		{Id: "link-1", Order: 0},
	})

	require.Empty(t, failures)
}

func Test_Unit_Link_UpdateOrder_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("missing-link").
		Return(nil, nil)

	failures := svc.Service.LinkService.UpdateOrder("user-uuid-test", false, []model.LinkOrderItem{
		{Id: "missing-link", Order: 0},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "missing-link", failures[0].Id)
	require.Contains(t, failures[0].Reason, "not found")
}

func Test_Unit_Link_UpdateOrder_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("link-1").
		Return(&model.Link{Id: "link-1", LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	failures := svc.Service.LinkService.UpdateOrder("someone-else-uuid-test", false, []model.LinkOrderItem{
		{Id: "link-1", Order: 0},
	})

	require.Len(t, failures, 1)
	require.Contains(t, failures[0].Reason, "forbidden")
}

func Test_Unit_Link_UpdateOrder_RepositoryFailure_ContinuesBatch(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("link-1").
		Return(&model.Link{Id: "link-1", LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.LinkRepository.
		EXPECT().
		UpdateOrder("link-1", 0).
		Return(errors.New("db unavailable"))

	svc.LinkRepository.
		EXPECT().
		Get("link-2").
		Return(&model.Link{Id: "link-2", LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.LinkRepository.
		EXPECT().
		UpdateOrder("link-2", 1).
		Return(nil)

	failures := svc.Service.LinkService.UpdateOrder("user-uuid-test", false, []model.LinkOrderItem{
		{Id: "link-1", Order: 0},
		{Id: "link-2", Order: 1},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "link-1", failures[0].Id)
}

func Test_Unit_Link_UpdateOrder_Success_Admin(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Get("link-1").
		Return(&model.Link{Id: "link-1", LinkBase: model.LinkBase{SectionId: "section-uuid-test"}}, nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	svc.LinkRepository.
		EXPECT().
		UpdateOrder("link-1", 0).
		Return(nil)

	failures := svc.Service.LinkService.UpdateOrder("admin-uuid-test", true, []model.LinkOrderItem{
		{Id: "link-1", Order: 0},
	})

	require.Empty(t, failures)
}
