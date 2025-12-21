package domain

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

func Test_Unit_Link_Create_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkRequest := &model.Link{
		Id: "link-uuid-test",
		LinkBase: model.LinkBase{
			Title:     "link-title-test",
			Link:      "https://example.com",
			Icon:      "my-base64-encoded-icon",
			Color:     "#FFFFFF",
			SectionId: "c5f3738e-668e-409e-bccd-c5c1b31de0da",
		},
	}

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
				Icon:      "my-base64-encoded-icon",
				Color:     "#FFFFFF",
				SectionId: "c5f3738e-668e-409e-bccd-c5c1b31de0da",
			},
		}, nil)

	link, err := svc.Service.LinkService.Create(linkRequest)

	require.NoError(t, err)
	require.NotNil(t, link)
	require.Equal(t, "link-uuid-test", link.Id)
	require.Equal(t, "link-title-test", link.Title)
}

func Test_Unit_Link_Create_Failure_Create(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Create(gomock.Any()).
		Return("", errors.New("an error occurred"))

	link, err := svc.Service.LinkService.Create(&model.Link{})

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, link)
}

func Test_Unit_Link_Create_Failure_Get(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Create(gomock.Any()).
		Return("link-uuid-test", nil)

	svc.LinkRepository.
		EXPECT().
		Get("link-uuid-test").
		Return(nil, errors.New("an error occurred"))

	link, err := svc.Service.LinkService.Create(&model.Link{})

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, link)
}

func Test_Unit_Link_Update_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	linkId := "link-uuid-test"

	updateRequest := &model.Link{
		Id: "link-uuid-test",
		LinkBase: model.LinkBase{
			Title:     "link-title-test-updated",
			Link:      "https://updated.example.com",
			Icon:      "my-base64-encoded-icon-updated",
			Color:     "#000000",
			SectionId: "c5f3738e-668e-409e-bccd-c5c1b31ma9da",
		},
	}

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
			Id: "link-uuid-test",
			LinkBase: model.LinkBase{
				Title:     "link-title-test-updated",
				Link:      "https://updated.example.com",
				Icon:      "my-base64-encoded-icon-updated",
				Color:     "#000000",
				SectionId: "c5f3738e-668e-409e-bccd-c5c1b31ma9da",
			},
		}, nil)

	link, err := svc.Service.LinkService.Update(linkId, updateRequest)

	require.NoError(t, err)
	require.NotNil(t, link)
	require.Equal(t, "link-title-test-updated", link.Title)
}

func Test_Unit_Link_Update_Failure_Update(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(errors.New("an error occurred"))

	link, err := svc.Service.LinkService.Update("link-uuid-test", &model.Link{})

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, link)
}
func Test_Unit_Link_Update_Failure_Get(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(nil)

	svc.LinkRepository.
		EXPECT().
		Get("link-uuid-test").
		Return(nil, errors.New("an error occurred"))

	link, err := svc.Service.LinkService.Update("link-uuid-test", &model.Link{})

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, link)
}

func Test_Unit_Link_Delete_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Delete(&model.Link{Id: "link-uuid-test"}).
		Return(nil)

	err := svc.Service.LinkService.Delete("link-uuid-test")

	require.NoError(t, err)
}

func Test_Unit_Link_Delete_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.LinkRepository.
		EXPECT().
		Delete(&model.Link{Id: "link-uuid-test"}).
		Return(errors.New("an error occurred"))

	err := svc.Service.LinkService.Delete("link-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
}
