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
