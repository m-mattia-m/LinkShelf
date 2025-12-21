package controller

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_API_CreateSection_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateSection(svc.Service)

	input := &model.SectionRequestBody{
		Body: model.SectionBase{
			Title:   "test-section",
			ShelfId: "shelf-uuid-test",
		},
	}

	svc.SectionService.
		EXPECT().
		Create(gomock.Any()).
		Return(&model.Section{
			Id: "section-uuid-test",
			SectionBase: model.SectionBase{
				Title:   "test-section",
				ShelfId: "shelf-uuid-test",
			},
		}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "section-uuid-test", resp.Body.Id)
	require.Equal(t, "test-section", resp.Body.Title)
}

func Test_API_CreateSection_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateSection(svc.Service)

	input := &model.SectionRequestBody{
		Body: model.SectionBase{
			Title:   "test-section",
			ShelfId: "shelf-uuid-test",
		},
	}

	svc.SectionService.
		EXPECT().
		Create(gomock.Any()).
		Return(nil, errors.New("failed to create section"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to create section")
}

func Test_API_GetSections_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetSections(svc.Service)

	svc.SectionService.
		EXPECT().
		List("shelf-uuid-test").
		Return([]model.Section{
			{
				Id: "section-uuid-test",
				SectionBase: model.SectionBase{
					Title:   "test-section",
					ShelfId: "shelf-uuid-test",
				},
			},
		}, nil)

	resp, err := handler(context.Background(), &model.SectionRequestShelfFilter{
		ShelfId: "shelf-uuid-test",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "section-uuid-test", resp.Body[0].Id)
	require.Equal(t, "test-section", resp.Body[0].Title)
}

func Test_API_GetSections_Failure_MissingShelfId(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetSections(svc.Service)

	resp, err := handler(context.Background(), &model.SectionRequestShelfFilter{
		ShelfId: "   ",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "shelfId is required")
}

func Test_API_GetSections_Failure_Service(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetSections(svc.Service)

	svc.SectionService.
		EXPECT().
		List("shelf-uuid-test").
		Return(nil, errors.New("failed to get section"))

	resp, err := handler(context.Background(), &model.SectionRequestShelfFilter{
		ShelfId: "shelf-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to get section")
}

func Test_API_UpdateSection_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateSection(svc.Service)

	input := &model.SectionRequestSectionFilterAndBody{
		SectionRequestSectionFilter: model.SectionRequestSectionFilter{
			SectionId: "section-uuid-test",
		},
		Body: model.SectionBase{
			Title:   "updated-section-title",
			ShelfId: "shelf-uuid-test",
		},
	}

	svc.SectionService.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(&model.Section{
			Id: "section-uuid-test",
			SectionBase: model.SectionBase{
				Title:   "updated-section-title",
				ShelfId: "shelf-uuid-test",
			},
		}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "section-uuid-test", resp.Body.Id)
	require.Equal(t, "updated-section-title", resp.Body.Title)
}

func Test_API_UpdateSection_Failure_MissingSectionId(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateSection(svc.Service)

	input := &model.SectionRequestSectionFilterAndBody{
		SectionRequestSectionFilter: model.SectionRequestSectionFilter{
			SectionId: "   ",
		},
		Body: model.SectionBase{
			Title: "updated-section-title",
		},
	}

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "sectionId is required")
}

func Test_API_UpdateSection_Failure_Service(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateSection(svc.Service)

	input := &model.SectionRequestSectionFilterAndBody{
		SectionRequestSectionFilter: model.SectionRequestSectionFilter{
			SectionId: "section-uuid-test",
		},
		Body: model.SectionBase{
			Title: "updated-section-title",
		},
	}

	svc.SectionService.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("failed to update section"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to update section")
}

func Test_API_DeleteSection_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteSection(svc.Service)

	svc.SectionService.
		EXPECT().
		Delete("section-uuid-test").
		Return(nil)

	resp, err := handler(context.Background(), &model.SectionRequestSectionFilter{
		SectionId: "section-uuid-test",
	})

	require.NoError(t, err)
	require.Nil(t, resp)
}

func Test_API_DeleteSection_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteSection(svc.Service)

	svc.SectionService.
		EXPECT().
		Delete("section-uuid-test").
		Return(errors.New("failed to delete section"))

	resp, err := handler(context.Background(), &model.SectionRequestSectionFilter{
		SectionId: "section-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to delete section")
}
