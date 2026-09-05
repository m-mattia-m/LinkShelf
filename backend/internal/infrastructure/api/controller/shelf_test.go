package controller

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_API_CreateShelf_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateShelf(svc.Service)

	input := &model.ShelfRequestBody{
		Body: model.ShelfBase{
			Title:  "test-shelf",
			UserId: "user-uuid-test",
		},
	}

	gomock.InOrder(
		svc.ShelfService.
			EXPECT().
			Create(gomock.Any()).
			Return("shelf-uuid-test", nil),

		svc.ShelfService.
			EXPECT().
			Get("shelf-uuid-test").
			Return(&model.Shelf{
				Id: "shelf-uuid-test",
				ShelfBase: model.ShelfBase{
					Title:  "test-shelf",
					UserId: "user-uuid-test",
				},
			}, nil),
	)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "shelf-uuid-test", resp.Body.Id)
	require.Equal(t, "test-shelf", resp.Body.Title)
}

func Test_API_CreateShelf_Failure_Create(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateShelf(svc.Service)

	input := &model.ShelfRequestBody{
		Body: model.ShelfBase{
			Title:  "test-shelf",
			UserId: "user-uuid-test",
		},
	}

	svc.ShelfService.
		EXPECT().
		Create(gomock.Any()).
		Return("", errors.New("failed to create user"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to create user")
}

func Test_API_CreateShelf_Failure_Get(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateShelf(svc.Service)

	input := &model.ShelfRequestBody{
		Body: model.ShelfBase{
			Title:  "test-shelf",
			UserId: "user-uuid-test",
		},
	}

	svc.ShelfService.
		EXPECT().
		Create(gomock.Any()).
		Return("shelf-uuid-test", nil)

	svc.ShelfService.
		EXPECT().
		Get("shelf-uuid-test").
		Return(nil, errors.New("failed to get user"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to get user")
}

func Test_API_GetShelfById_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetShelfById(svc.Service)

	svc.ShelfService.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{
			Id: "shelf-uuid-test",
			ShelfBase: model.ShelfBase{
				Title:  "test-shelf",
				UserId: "user-uuid-test",
			},
		}, nil)

	resp, err := handler(context.Background(), &model.ShelfRequestFilter{
		ShelfId: "shelf-uuid-test",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "shelf-uuid-test", resp.Body.Id)
	require.Equal(t, "test-shelf", resp.Body.Title)
}

func Test_API_GetShelfById_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetShelfById(svc.Service)

	svc.ShelfService.
		EXPECT().
		Get("shelf-uuid-test").
		Return(nil, errors.New("failed to get user"))

	resp, err := handler(context.Background(), &model.ShelfRequestFilter{
		ShelfId: "shelf-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to get user")
}

func Test_API_UpdateShelf_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateShelf(svc.Service)

	input := &model.ShelfFilterFilterAndBody{
		ShelfRequestFilter: model.ShelfRequestFilter{
			ShelfId: "shelf-uuid-test",
		},
		Body: model.ShelfBase{
			Title:  "updated-shelf-title",
			UserId: "user-uuid-test",
		},
	}

	svc.ShelfService.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(&model.Shelf{
			Id: "shelf-uuid-test",
			ShelfBase: model.ShelfBase{
				Title:  "updated-shelf-title",
				UserId: "user-uuid-test",
			},
		}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, "shelf-uuid-test", resp.Body.Id)
	require.Equal(t, "updated-shelf-title", resp.Body.Title)
}

func Test_API_UpdateShelf_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateShelf(svc.Service)

	input := &model.ShelfFilterFilterAndBody{
		ShelfRequestFilter: model.ShelfRequestFilter{
			ShelfId: "shelf-uuid-test",
		},
		Body: model.ShelfBase{
			Title: "updated-shelf-title",
		},
	}

	svc.ShelfService.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("failed to update user"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to update user")
}

func Test_API_DeleteShelf_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteShelf(svc.Service)

	gomock.InOrder(
		svc.ShelfService.
			EXPECT().
			Get("shelf-uuid-test").
			Return(&model.Shelf{
				Id: "shelf-uuid-test",
			}, nil),

		svc.ShelfService.
			EXPECT().
			Delete(gomock.Any()).
			Return(nil),
	)

	resp, err := handler(context.Background(), &model.ShelfRequestFilter{
		ShelfId: "shelf-uuid-test",
	})

	require.NoError(t, err)
	require.Nil(t, resp)
}

func Test_API_DeleteShelf_Failure_Get(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteShelf(svc.Service)

	svc.ShelfService.
		EXPECT().
		Get("shelf-uuid-test").
		Return(nil, errors.New("failed to get user"))

	resp, err := handler(context.Background(), &model.ShelfRequestFilter{
		ShelfId: "shelf-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to get user")
}

func Test_API_DeleteShelf_Failure_Delete(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteShelf(svc.Service)

	svc.ShelfService.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{
			Id: "shelf-uuid-test",
		}, nil)

	svc.ShelfService.
		EXPECT().
		Delete(gomock.Any()).
		Return(errors.New("failed to delete user"))

	resp, err := handler(context.Background(), &model.ShelfRequestFilter{
		ShelfId: "shelf-uuid-test",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to delete user")
}
