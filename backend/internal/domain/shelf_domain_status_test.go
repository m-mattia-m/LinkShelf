package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type recordingPublisher struct{ events []recordedEvent }

type recordedEvent struct {
	typ  string
	data shelfDomainChangedData
}

func (p *recordingPublisher) Publish(eventType string, data any) {
	p.events = append(p.events, recordedEvent{typ: eventType, data: data.(shelfDomainChangedData)})
}

func requireReady(t *testing.T, enabled bool) {
	t.Helper()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("customDomains.requireReady", enabled)
}

func Test_Unit_Shelf_RequireReady_ServesOnlyReadyDomains(t *testing.T) {
	t.Run("off: any saved domain is served", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		requireReady(t, false)

		svc.ShelfRepository.EXPECT().GetByDomain("profile.example.com").Return(&model.Shelf{}, nil)
		svc.ThemeRepository.EXPECT().Get(gomock.Any()).AnyTimes().Return(nil, nil)

		shelf, err := svc.Service.ShelfService.GetByDomain("profile.example.com")
		require.NoError(t, err)
		require.NotNil(t, shelf)
	})

	t.Run("on: only the ready lookup is used", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		requireReady(t, true)

		svc.ShelfRepository.EXPECT().GetReadyByDomain("profile.example.com").Return(nil, nil)

		shelf, err := svc.Service.ShelfService.GetByDomain("Profile.Example.com")
		require.NoError(t, err)
		require.Nil(t, shelf)
	})

	t.Run("registered for CORS follows the same rule", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()

		requireReady(t, false)
		svc.ShelfRepository.EXPECT().DomainInUse("profile.example.com", "").Return(true, nil)
		ok, err := svc.Service.ShelfService.IsDomainRegistered("profile.example.com")
		require.NoError(t, err)
		require.True(t, ok)

		requireReady(t, true)
		svc.ShelfRepository.EXPECT().DomainReadyInUse("profile.example.com", "").Return(false, nil)
		ok, err = svc.Service.ShelfService.IsDomainRegistered("profile.example.com")
		require.NoError(t, err)
		require.False(t, ok)
	})
}

func Test_Unit_Shelf_RequireReady_SavingAClaimedDomainIsNotAConflict(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	userBasedPaths(t, false)
	config.Set("customDomains.requireReady", true)

	svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
	// No DomainInUse call: the mock would fail the test on one.
	svc.ShelfRepository.EXPECT().Create(gomock.Any()).Return("shelf-1", nil)

	_, err := svc.Service.ShelfService.Create("user-1", shelfWithDomain("profile.example.com"))
	require.NoError(t, err)
}

func Test_Unit_Shelf_SetDomainStatus(t *testing.T) {
	status := model.DomainStatus{State: "pending", Reason: "txt_missing", Records: []model.DomainRecord{{Type: "TXT", Name: "n", Value: "v"}}}
	shelf := func() *model.Shelf {
		return &model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-1"}, UserId: "user-1", Domain: "profile.example.com"}
	}

	t.Run("stores the status and answers the item", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		requireReady(t, true)

		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(shelf(), nil)
		svc.ShelfRepository.EXPECT().SetDomainStatus("shelf-1", "profile.example.com", status, gomock.Any()).Return(true, nil)

		item, err := svc.Service.ShelfService.SetDomainStatus("Profile.Example.com", "shelf-1", status)
		require.NoError(t, err)
		require.Equal(t, "pending", item.Status.State)
		require.Equal(t, "profile.example.com", item.Status.ObservedDomain)
		require.Equal(t, "user-1", item.OwnerId)
		require.Len(t, item.Status.Records, 1)
	})

	t.Run("a stale controller gets a conflict", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		requireReady(t, true)

		changed := shelf()
		changed.Domain = "new.example.com"
		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(changed, nil)

		_, err := svc.Service.ShelfService.SetDomainStatus("profile.example.com", "shelf-1", status)
		require.ErrorIs(t, err, ErrConflict)
	})

	t.Run("a shelf which is gone is a conflict", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		requireReady(t, true)

		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(nil, nil)

		_, err := svc.Service.ShelfService.SetDomainStatus("profile.example.com", "shelf-1", status)
		require.ErrorIs(t, err, ErrConflict)
	})

	t.Run("the domain changing between read and write is a conflict", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		requireReady(t, true)

		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(shelf(), nil)
		svc.ShelfRepository.EXPECT().SetDomainStatus("shelf-1", "profile.example.com", status, gomock.Any()).Return(false, nil)

		_, err := svc.Service.ShelfService.SetDomainStatus("profile.example.com", "shelf-1", status)
		require.ErrorIs(t, err, ErrConflict)
	})

	t.Run("only one shelf can be ready on a domain", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		requireReady(t, true)

		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(shelf(), nil)
		svc.ShelfRepository.EXPECT().DomainReadyInUse("profile.example.com", "shelf-1").Return(true, nil)

		_, err := svc.Service.ShelfService.SetDomainStatus("profile.example.com", "shelf-1", model.DomainStatus{State: "ready"})
		require.ErrorIs(t, err, ErrConflict)
	})

	t.Run("an unknown state is invalid", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()

		_, err := svc.Service.ShelfService.SetDomainStatus("profile.example.com", "shelf-1", model.DomainStatus{State: "done"})
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("too many records are invalid", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()

		big := model.DomainStatus{State: "pending"}
		for i := 0; i < 10; i++ {
			big.Records = append(big.Records, model.DomainRecord{Type: "TXT", Name: "n", Value: string(make([]byte, 600))})
		}
		_, err := svc.Service.ShelfService.SetDomainStatus("profile.example.com", "shelf-1", big)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("database errors are returned", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		requireReady(t, true)

		boom := errors.New("boom")
		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(nil, boom)
		_, err := svc.Service.ShelfService.SetDomainStatus("profile.example.com", "shelf-1", status)
		require.ErrorIs(t, err, boom)
	})
}

func Test_Unit_Shelf_ListDomains_Paginates(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	id1, id2, id3 := uuid.Must(uuid.NewV7()).String(), uuid.Must(uuid.NewV7()).String(), uuid.Must(uuid.NewV7()).String()
	rows := []model.Shelf{
		{PublicShelf: model.PublicShelf{Id: id1}, UserId: "u", Domain: "a.example.com"},
		{PublicShelf: model.PublicShelf{Id: id2}, UserId: "u", Domain: "b.example.com", DomainStatus: &model.DomainStatus{State: "ready"}, DomainUpdatedAt: 4102444800},
		{PublicShelf: model.PublicShelf{Id: id3}, UserId: "u", Domain: "c.example.com"},
	}
	// One more than asked for, to learn whether a next page exists.
	svc.ShelfRepository.EXPECT().ListDomains("", 3).Return(rows, nil)
	svc.ShelfRepository.EXPECT().ListDomains(id2, 3).Return(rows[2:], nil)

	items, next, err := svc.Service.ShelfService.ListDomains("", 2)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, id2, next)
	require.Equal(t, "", items[0].Status.State, "nothing reported yet")
	require.Equal(t, "ready", items[1].Status.State)
	require.Equal(t, "2100-01-01T00:00:00Z", items[1].UpdatedAt)
	require.NotEqual(t, "1970-01-01T00:00:00Z", items[0].CreatedAt, "taken from the UUIDv7")
	require.NotNil(t, items[0].Status.Records)

	items, next, err = svc.Service.ShelfService.ListDomains(next, 2)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Empty(t, next)
}

func Test_Unit_Shelf_PublishesDomainChanges(t *testing.T) {
	t.Run("create, change, clear and delete", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)
		events := &recordingPublisher{}
		svc.Service.Events = events

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil).AnyTimes()
		svc.ThemeRepository.EXPECT().Get(gomock.Any()).AnyTimes().Return(nil, nil)
		svc.ShelfRepository.EXPECT().DomainInUse(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
		svc.ShelfRepository.EXPECT().PathInUse(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
		svc.ShelfRepository.EXPECT().Create(gomock.Any()).Return("shelf-1", nil)
		svc.ShelfRepository.EXPECT().Update(gomock.Any()).AnyTimes().Return(nil)
		svc.ShelfRepository.EXPECT().Delete(gomock.Any()).Return(nil)
		current := &model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-1"}, UserId: "user-1", Domain: "a.example.com"}
		svc.ShelfRepository.EXPECT().Get("shelf-1").AnyTimes().DoAndReturn(func(string) (*model.Shelf, error) { c := *current; return &c, nil })

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithDomain("a.example.com"))
		require.NoError(t, err)

		// Same domain again: nothing to tell.
		_, err = svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithDomain("a.example.com"))
		require.NoError(t, err)

		_, err = svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithDomain("b.example.com"))
		require.NoError(t, err)

		// A path instead of a domain clears it.
		_, err = svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithPath("my-path"))
		require.NoError(t, err)

		require.NoError(t, svc.Service.ShelfService.Delete(current))

		require.Equal(t, []recordedEvent{
			{EventShelfDomainChanged, shelfDomainChangedData{ShelfId: "shelf-1", OwnerId: "user-1", Domain: "a.example.com"}},
			{EventShelfDomainChanged, shelfDomainChangedData{ShelfId: "shelf-1", OwnerId: "user-1", Domain: "b.example.com", PreviousDomain: "a.example.com"}},
			{EventShelfDomainChanged, shelfDomainChangedData{ShelfId: "shelf-1", OwnerId: "user-1", Domain: "", PreviousDomain: "a.example.com"}},
			{EventShelfDomainChanged, shelfDomainChangedData{ShelfId: "shelf-1", OwnerId: "user-1", Domain: "", PreviousDomain: "a.example.com"}},
		}, events.events)
	})

	t.Run("a shelf with a path only says nothing", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		events := &recordingPublisher{}
		svc.Service.Events = events

		svc.ShelfRepository.EXPECT().Delete(gomock.Any()).Return(nil)
		require.NoError(t, svc.Service.ShelfService.Delete(&model.Shelf{PublicShelf: model.PublicShelf{Id: "s", Path: "p"}, UserId: "u"}))
		require.Empty(t, events.events)
	})

	t.Run("deleting a user reports the domains of its shelves", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		events := &recordingPublisher{}
		svc.Service.Events = events

		svc.ShelfRepository.EXPECT().ListByUserId("user-1").Return([]model.Shelf{
			{PublicShelf: model.PublicShelf{Id: "s-1"}, UserId: "user-1", Domain: "a.example.com"},
			{PublicShelf: model.PublicShelf{Id: "s-2", Path: "p"}, UserId: "user-1"},
		}, nil)
		svc.UserRepository.EXPECT().Delete(gomock.Any()).Return(nil)

		require.NoError(t, svc.Service.UserService.Delete(&model.User{Id: "user-1"}))
		require.Len(t, events.events, 1)
		require.Equal(t, "a.example.com", events.events[0].data.PreviousDomain)
	})

	t.Run("a failed delete reports nothing", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		events := &recordingPublisher{}
		svc.Service.Events = events

		svc.ShelfRepository.EXPECT().ListByUserId("user-1").Return([]model.Shelf{{Domain: "a.example.com"}}, nil)
		svc.UserRepository.EXPECT().Delete(gomock.Any()).Return(errors.New("boom"))

		require.Error(t, svc.Service.UserService.Delete(&model.User{Id: "user-1"}))
		require.Empty(t, events.events)
	})
}
