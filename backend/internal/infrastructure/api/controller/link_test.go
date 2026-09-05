package controller

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_API_CreateLink_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateLink(svc.Service)

	input := &model.LinkRequestBody{
		Body: model.LinkBase{
			Title:     "test-link",
			Color:     "#ff0000",
			SectionId: "section-uuid-test",
		},
	}

	svc.LinkService.
		EXPECT().
		Create(gomock.Any()).
		Return(&model.Link{
			Id: "link-uuid-test",
			LinkBase: model.LinkBase{
				Title:     "test-link",
				SectionId: "shelf-uuid-test",
				Color:     "#ff0000",
			},
		}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "link-uuid-test", resp.Body.Id)
	require.Equal(t, "test-link", resp.Body.Title)
}

func Test_API_CreateLink_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateLink(svc.Service)
	input := &model.LinkRequestBody{
		Body: model.LinkBase{
			Title:     "test-link",
			Color:     "#ff0000",
			SectionId: "section-uuid-test",
		},
	}

	svc.LinkService.
		EXPECT().
		Create(gomock.Any()).
		Return(nil, errors.New("failed to create link"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to create link")
}

func Test_API_GetLinks_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetLinks(svc.Service)

	svc.LinkService.
		EXPECT().
		List(gomock.Any()).
		Return([]model.Link{
			{
				Id: "link-uuid-test",
				LinkBase: model.LinkBase{
					Title:     "test-link",
					Link:      "https://example.com",
					Icon:      "icon-test",
					Color:     "#ff0000",
					SectionId: "section-uuid-test",
				},
			},
		}, nil)

	resp, err := handler(context.Background(), &model.LinkRequestShelfFilter{
		ShelfId: "link-uuid-test",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "link-uuid-test", resp.Body[0].Id)
	require.Equal(t, "test-link", resp.Body[0].Title)
	require.Equal(t, "https://example.com", resp.Body[0].Link)
	require.Equal(t, "icon-test", resp.Body[0].Icon)
	require.Equal(t, "#ff0000", resp.Body[0].Color)
}

func Test_API_GetLinks_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetLinks(svc.Service)

	svc.LinkService.
		EXPECT().
		List(gomock.Any()).
		Return(nil, errors.New("failed to get links"))

	resp, err := handler(context.Background(), &model.LinkRequestShelfFilter{
		ShelfId: "link-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to get links")
}

func Test_API_UpdateLink_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateLink(svc.Service)

	input := &model.LinkFilterFilterAndBody{
		LinkRequestFilter: model.LinkRequestFilter{
			LinkId: "link-uuid-test",
		},
		Body: model.LinkBase{
			Title:     "updated-link-title",
			Color:     "#00ff00",
			SectionId: "section-uuid-test",
		},
	}

	svc.LinkService.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(&model.Link{
			Id: "link-uuid-test",
			LinkBase: model.LinkBase{
				Title:     "updated-link-title",
				Color:     "#00ff00",
				SectionId: "section-uuid-test",
			},
		}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "link-uuid-test", resp.Body.Id)
	require.Equal(t, "updated-link-title", resp.Body.Title)
	require.Equal(t, "#00ff00", resp.Body.Color)
}

func Test_API_UpdateLink_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateLink(svc.Service)

	input := &model.LinkFilterFilterAndBody{
		LinkRequestFilter: model.LinkRequestFilter{
			LinkId: "link-uuid-test",
		},
		Body: model.LinkBase{
			Title:     "updated-link-title",
			Color:     "#00ff00",
			SectionId: "section-uuid-test",
		},
	}

	svc.LinkService.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("failed to update link"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to update link")
}

func Test_API_DeleteLink_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteLink(svc.Service)

	svc.LinkService.
		EXPECT().
		Delete(gomock.Any()).
		Return(nil)

	resp, err := handler(context.Background(), &model.LinkRequestFilter{
		LinkId: "link-uuid-test",
	})

	require.NoError(t, err)
	require.Nil(t, resp)
}

func Test_API_DeleteLink_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteLink(svc.Service)

	svc.LinkService.
		EXPECT().
		Delete(gomock.Any()).
		Return(errors.New("failed to delete link"))

	resp, err := handler(context.Background(), &model.LinkRequestFilter{
		LinkId: "link-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to delete link")
}
