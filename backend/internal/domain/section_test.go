package domain

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Unit_Section_List_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	sections := []model.Section{
		{Id: "section-1"},
		{Id: "section-2"},
	}

	svc.SectionRepository.
		EXPECT().
		ListByShelfId("shelf-uuid-test").
		Return(sections, nil)

	result, err := svc.Service.SectionService.List("shelf-uuid-test")

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "section-1", result[0].Id)
}

func Test_Unit_Section_List_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		ListByShelfId("shelf-uuid-test").
		Return(nil, errors.New("an error occurred"))

	sections, err := svc.Service.SectionService.List("shelf-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, sections)
}

func Test_Unit_Section_Get_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{
			Id: "section-uuid-test",
		}, nil)

	section, err := svc.Service.SectionService.Get("section-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, section)
	require.Equal(t, "section-uuid-test", section.Id)
}

func Test_Unit_Section_Get_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(nil, errors.New("an error occurred"))

	section, err := svc.Service.SectionService.Get("section-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, section)
}

func Test_Unit_Section_Create_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	sectionRequest := &model.Section{
		SectionBase: model.SectionBase{
			Title:   "section-title-test",
			ShelfId: "shelf-uuid-test",
		},
	}

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		Create(sectionRequest).
		Return("section-uuid-test", nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{
			Id: "section-uuid-test",
			SectionBase: model.SectionBase{
				Title:   "section-title-test",
				ShelfId: "shelf-uuid-test",
			},
		}, nil)

	section, err := svc.Service.SectionService.Create("user-uuid-test", false, sectionRequest)

	require.NoError(t, err)
	require.NotNil(t, section)
	require.Equal(t, "section-uuid-test", section.Id)
	require.Equal(t, "section-title-test", section.Title)
}

func Test_Unit_Section_Create_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	sectionRequest := &model.Section{
		SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"},
	}

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	section, err := svc.Service.SectionService.Create("someone-else-uuid-test", false, sectionRequest)

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, section)
}

func Test_Unit_Section_Create_NotFound_Shelf(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	sectionRequest := &model.Section{
		SectionBase: model.SectionBase{ShelfId: "missing-shelf"},
	}

	svc.ShelfRepository.
		EXPECT().
		Get("missing-shelf").
		Return(nil, nil)

	section, err := svc.Service.SectionService.Create("user-uuid-test", false, sectionRequest)

	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, section)
}

func Test_Unit_Section_Create_Failure_Create(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	sectionRequest := &model.Section{SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		Create(gomock.Any()).
		Return("", errors.New("an error occurred"))

	section, err := svc.Service.SectionService.Create("user-uuid-test", false, sectionRequest)

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, section)
}

func Test_Unit_Section_Update_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	sectionId := "section-uuid-test"

	updateRequest := &model.Section{
		SectionBase: model.SectionBase{
			Title:   "updated-title",
			ShelfId: "shelf-uuid-test",
		},
	}

	svc.SectionRepository.
		EXPECT().
		Get(sectionId).
		Return(&model.Section{Id: sectionId, SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		Update(&model.Section{
			Id:          sectionId,
			SectionBase: updateRequest.SectionBase,
		}).
		Return(nil)

	svc.SectionRepository.
		EXPECT().
		Get(sectionId).
		Return(&model.Section{
			Id: sectionId,
			SectionBase: model.SectionBase{
				Title:   "updated-title",
				ShelfId: "shelf-uuid-test",
			},
		}, nil)

	section, err := svc.Service.SectionService.Update(sectionId, "user-uuid-test", false, updateRequest)

	require.NoError(t, err)
	require.NotNil(t, section)
	require.Equal(t, "updated-title", section.Title)
}

func Test_Unit_Section_Update_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	sectionId := "section-uuid-test"

	svc.SectionRepository.
		EXPECT().
		Get(sectionId).
		Return(&model.Section{Id: sectionId, SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	section, err := svc.Service.SectionService.Update(sectionId, "someone-else-uuid-test", false, &model.Section{})

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, section)
}

func Test_Unit_Section_Update_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(nil, nil)

	section, err := svc.Service.SectionService.Update("section-uuid-test", "user-uuid-test", false, &model.Section{})

	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, section)
}

func Test_Unit_Section_Delete_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		Delete(&model.Section{Id: "section-uuid-test"}).
		Return(nil)

	err := svc.Service.SectionService.Delete("section-uuid-test", "user-uuid-test", false)

	require.NoError(t, err)
}

func Test_Unit_Section_Delete_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-uuid-test").
		Return(&model.Section{Id: "section-uuid-test", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	err := svc.Service.SectionService.Delete("section-uuid-test", "someone-else-uuid-test", false)

	require.ErrorIs(t, err, ErrForbidden)
}

func Test_Unit_Section_UpdateOrder_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-1").
		Return(&model.Section{Id: "section-1", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		UpdateOrder("section-1", 0).
		Return(nil)

	svc.SectionRepository.
		EXPECT().
		Get("section-2").
		Return(&model.Section{Id: "section-2", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		UpdateOrder("section-2", 1).
		Return(nil)

	failures := svc.Service.SectionService.UpdateOrder("user-uuid-test", false, []model.SectionOrderItem{
		{Id: "section-1", Order: 0},
		{Id: "section-2", Order: 1},
	})

	require.Empty(t, failures)
}

func Test_Unit_Section_UpdateOrder_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("missing-section").
		Return(nil, nil)

	failures := svc.Service.SectionService.UpdateOrder("user-uuid-test", false, []model.SectionOrderItem{
		{Id: "missing-section", Order: 0},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "missing-section", failures[0].Id)
	require.Contains(t, failures[0].Reason, "not found")
}

func Test_Unit_Section_UpdateOrder_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-1").
		Return(&model.Section{Id: "section-1", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	failures := svc.Service.SectionService.UpdateOrder("someone-else-uuid-test", false, []model.SectionOrderItem{
		{Id: "section-1", Order: 0},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "section-1", failures[0].Id)
	require.Contains(t, failures[0].Reason, "forbidden")
}

func Test_Unit_Section_UpdateOrder_RepositoryFailure_ContinuesBatch(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-1").
		Return(&model.Section{Id: "section-1", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		UpdateOrder("section-1", 0).
		Return(errors.New("db unavailable"))

	svc.SectionRepository.
		EXPECT().
		Get("section-2").
		Return(&model.Section{Id: "section-2", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "user-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		UpdateOrder("section-2", 1).
		Return(nil)

	failures := svc.Service.SectionService.UpdateOrder("user-uuid-test", false, []model.SectionOrderItem{
		{Id: "section-1", Order: 0},
		{Id: "section-2", Order: 1},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "section-1", failures[0].Id)
	require.Contains(t, failures[0].Reason, "db unavailable")
}

func Test_Unit_Section_UpdateOrder_Success_Admin(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SectionRepository.
		EXPECT().
		Get("section-1").
		Return(&model.Section{Id: "section-1", SectionBase: model.SectionBase{ShelfId: "shelf-uuid-test"}}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"}, UserId: "owner-uuid-test"}, nil)

	svc.SectionRepository.
		EXPECT().
		UpdateOrder("section-1", 0).
		Return(nil)

	failures := svc.Service.SectionService.UpdateOrder("admin-uuid-test", true, []model.SectionOrderItem{
		{Id: "section-1", Order: 0},
	})

	require.Empty(t, failures)
}
