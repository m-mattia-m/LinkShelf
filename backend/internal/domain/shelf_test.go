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
	}

	svc.ShelfRepository.
		EXPECT().
		Create(&model.Shelf{
			PublicShelf: model.PublicShelf{Title: "shelf-title-test"},
			UserId:      "user-uuid-test",
		}).
		Return("shelf-uuid-test", nil)

	shelfId, err := svc.Service.ShelfService.Create("user-uuid-test", shelfRequest)

	require.NoError(t, err)
	require.Equal(t, "shelf-uuid-test", shelfId)
}

func Test_Unit_Shelf_Creation_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{Title: "shelf-title-test"},
	}

	svc.ShelfRepository.
		EXPECT().
		Create(gomock.Any()).
		Return("", errors.New("an error occurred"))

	shelfId, err := svc.Service.ShelfService.Create("user-uuid-test", shelfRequest)

	require.ErrorContains(t, err, "an error occurred")
	require.Empty(t, shelfId)
}

func Test_Unit_Shelf_Update_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfId := "shelf-uuid-test"

	updateRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{Title: "updated-title"},
	}

	svc.ShelfRepository.
		EXPECT().
		Get(shelfId).
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: shelfId},
			UserId:      "user-uuid-test",
		}, nil)

	svc.ShelfRepository.
		EXPECT().
		Update(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: shelfId, Title: "updated-title"},
		}).
		Return(nil)

	svc.ShelfRepository.
		EXPECT().
		Get(shelfId).
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: shelfId, Title: "updated-title"},
			UserId:      "user-uuid-test",
		}, nil)

	shelf, err := svc.Service.ShelfService.Update(shelfId, "user-uuid-test", false, updateRequest)

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, shelfId, shelf.Id)
	require.Equal(t, "updated-title", shelf.Title)
}

func Test_Unit_Shelf_Update_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfId := "shelf-uuid-test"

	svc.ShelfRepository.
		EXPECT().
		Get(shelfId).
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: shelfId},
			UserId:      "owner-uuid-test",
		}, nil)

	shelf, err := svc.Service.ShelfService.Update(shelfId, "someone-else-uuid-test", false, &model.Shelf{})

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Update_Success_Admin_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfId := "shelf-uuid-test"

	svc.ShelfRepository.
		EXPECT().
		Get(shelfId).
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: shelfId},
			UserId:      "owner-uuid-test",
		}, nil)

	svc.ShelfRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(nil)

	svc.ShelfRepository.
		EXPECT().
		Get(shelfId).
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: shelfId, Title: "updated-by-admin"},
			UserId:      "owner-uuid-test",
		}, nil)

	shelf, err := svc.Service.ShelfService.Update(shelfId, "admin-uuid-test", true, &model.Shelf{})

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "updated-by-admin", shelf.Title)
}

func Test_Unit_Shelf_Update_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(nil, nil)

	shelf, err := svc.Service.ShelfService.Update("shelf-uuid-test", "user-uuid-test", false, &model.Shelf{})

	require.NoError(t, err)
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Update_Failure_Get(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(nil, errors.New("an error occurred"))

	shelf, err := svc.Service.ShelfService.Update("shelf-uuid-test", "user-uuid-test", false, &model.Shelf{})

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Get_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"},
			UserId:      "user-uuid-test",
		}, nil)

	shelf, err := svc.Service.ShelfService.Get("shelf-uuid-test", "user-uuid-test", false)

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "shelf-uuid-test", shelf.Id)
}

func Test_Unit_Shelf_Get_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"},
			UserId:      "owner-uuid-test",
		}, nil)

	shelf, err := svc.Service.ShelfService.Get("shelf-uuid-test", "someone-else-uuid-test", false)

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Get_Success_Admin(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"},
			UserId:      "owner-uuid-test",
		}, nil)

	shelf, err := svc.Service.ShelfService.Get("shelf-uuid-test", "admin-uuid-test", true)

	require.NoError(t, err)
	require.NotNil(t, shelf)
}

func Test_Unit_Shelf_Get_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(nil, errors.New("an error occurred"))

	shelf, err := svc.Service.ShelfService.Get("shelf-uuid-test", "user-uuid-test", false)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_List_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		ListByUserId("user-uuid-test").
		Return([]model.Shelf{{PublicShelf: model.PublicShelf{Id: "shelf-1"}}}, nil)

	shelves, err := svc.Service.ShelfService.List("user-uuid-test", false)

	require.NoError(t, err)
	require.Len(t, shelves, 1)
}

func Test_Unit_Shelf_List_Admin(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		List().
		Return([]model.Shelf{{PublicShelf: model.PublicShelf{Id: "shelf-1"}}, {PublicShelf: model.PublicShelf{Id: "shelf-2"}}}, nil)

	shelves, err := svc.Service.ShelfService.List("admin-uuid-test", true)

	require.NoError(t, err)
	require.Len(t, shelves, 2)
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
