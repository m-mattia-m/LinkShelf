package domain

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Unit_Shelf_Creation_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title: "shelf-title-test",
		},
		UserId: "user-uuid-test",
	}

	svc.ShelfRepository.
		EXPECT().
		Create(shelfRequest).
		Return("shelf-uuid-test", nil)

	shelfId, err := svc.Service.ShelfService.Create(shelfRequest)

	require.NoError(t, err)
	require.Equal(t, "shelf-uuid-test", shelfId)
}

func Test_Unit_Shelf_Creation_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title: "shelf-title-test",
		},
		UserId: "user-uuid-test",
	}

	svc.ShelfRepository.
		EXPECT().
		Create(shelfRequest).
		Return("", errors.New("an error occurred"))

	shelfId, err := svc.Service.ShelfService.Create(shelfRequest)

	require.ErrorContains(t, err, "an error occurred")
	require.Empty(t, shelfId)
}

func Test_Unit_Shelf_Update_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfId := "shelf-uuid-test"

	updateRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title: "updated-title",
		},
		UserId: "user-uuid-test",
	}

	// Update must be called with ID set by service
	svc.ShelfRepository.
		EXPECT().
		Update(&model.Shelf{
			PublicShelf: model.PublicShelf{
				Id:    shelfId,
				Title: "updated-title",
			},
			UserId: "user-uuid-test",
		}).
		Return(nil)

	svc.ShelfRepository.
		EXPECT().
		Get(shelfId).
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{
				Id:    shelfId,
				Title: "updated-title",
			},
			UserId: "user-uuid-test",
		}, nil)

	shelf, err := svc.Service.ShelfService.Update(shelfId, updateRequest)

	require.NoError(t, err)
	require.NotNil(t, shelf)

	require.Equal(t, shelfId, shelf.Id)
	require.Equal(t, "updated-title", shelf.Title)
	require.Equal(t, "user-uuid-test", shelf.UserId)
}

func Test_Unit_Shelf_Update_Failure_Update(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfId := "shelf-uuid-test"

	updateRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title: "updated-title",
		},
		UserId: "user-uuid-test",
	}

	svc.ShelfRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(errors.New("an error occurred"))

	shelf, err := svc.Service.ShelfService.Update(shelfId, updateRequest)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Update_Failure_Get(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfId := "shelf-uuid-test"

	updateRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title: "updated-title",
		},
		UserId: "user-uuid-test",
	}

	svc.ShelfRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(nil)

	svc.ShelfRepository.
		EXPECT().
		Get(shelfId).
		Return(nil, errors.New("an error occurred"))

	shelf, err := svc.Service.ShelfService.Update(shelfId, updateRequest)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Get_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{
				Id: "shelf-uuid-test",
			},
		}, nil)

	shelf, err := svc.Service.ShelfService.Get("shelf-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "shelf-uuid-test", shelf.Id)
}

func Test_Unit_Shelf_Get_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(nil, errors.New("an error occurred"))

	shelf, err := svc.Service.ShelfService.Get("shelf-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_GetByPath_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		GetByPath("my-path").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{
				Id:   "shelf-uuid-test",
				Path: "my-path",
			},
		}, nil)

	shelf, err := svc.Service.ShelfService.GetByPath("my-path")

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "shelf-uuid-test", shelf.Id)
}

func Test_Unit_Shelf_GetByPath_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		GetByPath("missing-path").
		Return(nil, nil)

	shelf, err := svc.Service.ShelfService.GetByPath("missing-path")

	require.NoError(t, err)
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_GetByPath_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		GetByPath("my-path").
		Return(nil, errors.New("an error occurred"))

	shelf, err := svc.Service.ShelfService.GetByPath("my-path")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Delete_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelf := &model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}}

	svc.ShelfRepository.
		EXPECT().
		Delete(shelf).
		Return(nil)

	err := svc.Service.ShelfService.Delete(shelf)

	require.NoError(t, err)
}

func Test_Unit_Shelf_Delete_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelf := &model.Shelf{
		PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"},
	}

	svc.ShelfRepository.
		EXPECT().
		Delete(shelf).
		Return(errors.New("an error occurred"))

	err := svc.Service.ShelfService.Delete(shelf)

	require.ErrorContains(t, err, "an error occurred")
}
