package domain

import (
	"backend/internal/config"
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
			Path:  "shelf-path-test",
		},
	}

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

	svc.ShelfRepository.
		EXPECT().
		PathInUse("shelf-path-test", "").
		Return(false, nil)

	svc.ShelfRepository.
		EXPECT().
		Create(&model.Shelf{
			PublicShelf: model.PublicShelf{Title: "shelf-title-test", Path: "shelf-path-test"},
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
		PublicShelf: model.PublicShelf{Title: "shelf-title-test", Path: "shelf-path-test"},
	}

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

	svc.ShelfRepository.
		EXPECT().
		PathInUse("shelf-path-test", "").
		Return(false, nil)

	svc.ShelfRepository.
		EXPECT().
		Create(gomock.Any()).
		Return("", errors.New("an error occurred"))

	shelfId, err := svc.Service.ShelfService.Create("user-uuid-test", shelfRequest)

	require.ErrorContains(t, err, "an error occurred")
	require.Empty(t, shelfId)
}

func Test_Unit_Shelf_Creation_UserNotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, nil)

	shelfId, err := svc.Service.ShelfService.Create("user-uuid-test", &model.Shelf{})

	require.ErrorIs(t, err, ErrNotFound)
	require.Empty(t, shelfId)
}

func Test_Unit_Shelf_Creation_UserLookupFailure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	shelfId, err := svc.Service.ShelfService.Create("user-uuid-test", &model.Shelf{})

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

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

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

	svc.UserRepository.
		EXPECT().
		Get("someone-else-uuid-test").
		Return(&model.User{Id: "someone-else-uuid-test"}, nil)

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

	svc.UserRepository.
		EXPECT().
		Get("admin-uuid-test").
		Return(&model.User{Id: "admin-uuid-test"}, nil)

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

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

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

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(nil, errors.New("an error occurred"))

	shelf, err := svc.Service.ShelfService.Update("shelf-uuid-test", "user-uuid-test", false, &model.Shelf{})

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Update_UserNotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(nil, nil)

	shelf, err := svc.Service.ShelfService.Update("shelf-uuid-test", "user-uuid-test", false, &model.Shelf{})

	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Update_UserLookupFailure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
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

func Test_Unit_Shelf_Creation_ThemeAssignable(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{Title: "shelf-title-test", Path: "shelf-path-test"},
		ThemeId:     "theme-uuid-test",
	}

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test"}, nil)

	svc.ShelfRepository.
		EXPECT().
		PathInUse("shelf-path-test", "").
		Return(false, nil)

	svc.ShelfRepository.
		EXPECT().
		Create(shelfRequest).
		Return("shelf-uuid-test", nil)

	shelfId, err := svc.Service.ShelfService.Create("user-uuid-test", shelfRequest)

	require.NoError(t, err)
	require.Equal(t, "shelf-uuid-test", shelfId)
}

func Test_Unit_Shelf_Creation_ThemeNotAssignable(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfRequest := &model.Shelf{
		PublicShelf: model.PublicShelf{Title: "shelf-title-test"},
		ThemeId:     "theme-uuid-test",
	}

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "someone-else-uuid-test"}, nil)

	shelfId, err := svc.Service.ShelfService.Create("user-uuid-test", shelfRequest)

	require.ErrorIs(t, err, ErrForbidden)
	require.Empty(t, shelfId)
}

func Test_Unit_Shelf_Update_ThemeNotAssignable(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	shelfId := "shelf-uuid-test"

	svc.UserRepository.
		EXPECT().
		Get("user-uuid-test").
		Return(&model.User{Id: "user-uuid-test"}, nil)

	svc.ShelfRepository.
		EXPECT().
		Get(shelfId).
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: shelfId},
			UserId:      "user-uuid-test",
		}, nil)

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, nil)

	shelf, err := svc.Service.ShelfService.Update(shelfId, "user-uuid-test", false, &model.Shelf{ThemeId: "theme-uuid-test"})

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_Get_AnnotatesThemeMissing(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"},
			UserId:      "user-uuid-test",
			ThemeId:     "deleted-theme-uuid",
		}, nil)

	svc.ThemeRepository.
		EXPECT().
		Get("deleted-theme-uuid").
		Return(nil, nil)

	shelf, err := svc.Service.ShelfService.Get("shelf-uuid-test", "user-uuid-test", false)

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.True(t, shelf.ThemeMissing)
}

func Test_Unit_Shelf_Get_ThemeStillPresent_NotMissing(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		Get("shelf-uuid-test").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-uuid-test"},
			UserId:      "user-uuid-test",
			ThemeId:     "theme-uuid-test",
		}, nil)

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Config: "--shelf-bg: #1c274c;\n"}, nil)

	shelf, err := svc.Service.ShelfService.Get("shelf-uuid-test", "user-uuid-test", false)

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.False(t, shelf.ThemeMissing)
}

func Test_Unit_Shelf_GetByPath_ResolvesThemeConfig(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		GetByPath("my-path").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-uuid-test", Path: "my-path"},
			ThemeId:     "theme-uuid-test",
		}, nil)

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Config: "--shelf-bg: #1c274c;\n"}, nil)

	shelf, err := svc.Service.ShelfService.GetByPath("my-path")

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Equal(t, "#1c274c", shelf.Theme["--shelf-bg"])
}

func Test_Unit_Shelf_GetByPath_MissingThemeFallsBackSilently(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfRepository.
		EXPECT().
		GetByPath("my-path").
		Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-uuid-test", Path: "my-path"},
			ThemeId:     "deleted-theme-uuid",
		}, nil)

	svc.ThemeRepository.
		EXPECT().
		Get("deleted-theme-uuid").
		Return(nil, nil)

	shelf, err := svc.Service.ShelfService.GetByPath("my-path")

	require.NoError(t, err)
	require.NotNil(t, shelf)
	require.Nil(t, shelf.Theme)
}

func userBasedPaths(t *testing.T, enabled bool) {
	t.Helper()
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("app.userBasedPaths", enabled)
}

func shelfWithPath(path string) *model.Shelf {
	return &model.Shelf{PublicShelf: model.PublicShelf{Title: "shelf-title-test", Path: path}}
}

func shelfWithDomain(domain string) *model.Shelf {
	return &model.Shelf{PublicShelf: model.PublicShelf{Title: "shelf-title-test"}, Domain: domain}
}

func Test_Unit_Shelf_Creation_PathRules_UserBasedPathsOff(t *testing.T) {
	t.Run("a free path is accepted", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().PathInUse("my-path", "").Return(false, nil)
		svc.ShelfRepository.EXPECT().Create(gomock.Any()).Return("shelf-1", nil)

		id, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("my-path"))

		require.NoError(t, err)
		require.Equal(t, "shelf-1", id)
	})

	t.Run("a path used by anyone else is a conflict", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().PathInUse("profile", "").Return(true, nil)

		id, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("profile"))

		require.ErrorIs(t, err, ErrConflict)
		require.Empty(t, id)
	})

	t.Run("a route word is rejected", func(t *testing.T) {
		for _, path := range []string{"app", "auth", "docs", "cloud", "about", "contact", "imprint", "privacy-policy", "terms-of-use", "api", "v1", "swagger", "health", "images", "DOCS"} {
			svc := NewMockService(t)
			userBasedPaths(t, false)

			svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)

			id, err := svc.Service.ShelfService.Create("user-1", shelfWithPath(path))

			require.ErrorIs(t, err, ErrInvalidInput, path)
			require.ErrorContains(t, err, "reserved", path)
			require.Empty(t, id)
			svc.Ctrl.Finish()
		}
	})

	t.Run("words that only usernames may not use are fine as a path", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().PathInUse("support", "").Return(false, nil)
		svc.ShelfRepository.EXPECT().Create(gomock.Any()).Return("shelf-1", nil)

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("support"))

		require.NoError(t, err)
	})

	t.Run("a shelf with neither a path nor a domain is rejected", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)

		id, err := svc.Service.ShelfService.Create("user-1", shelfWithPath(""))

		require.ErrorIs(t, err, ErrInvalidInput)
		require.ErrorContains(t, err, "path or a domain")
		require.Empty(t, id)
	})

	t.Run("a failing lookup is returned", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().PathInUse("my-path", "").Return(false, errors.New("db unavailable"))

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("my-path"))

		require.ErrorContains(t, err, "db unavailable")
	})
}

func Test_Unit_Shelf_Creation_PathRules_UserBasedPathsOn(t *testing.T) {
	t.Run("a path only has to be unique for the owner", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1", UserBase: model.UserBase{Username: "alice"}}, nil)
		svc.ShelfRepository.EXPECT().PathInUseByUser("user-1", "profile", "").Return(false, nil)
		svc.ShelfRepository.EXPECT().Create(gomock.Any()).Return("shelf-1", nil)

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("profile"))

		require.NoError(t, err)
	})

	t.Run("a second shelf with the same path for the same owner conflicts", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1", UserBase: model.UserBase{Username: "alice"}}, nil)
		svc.ShelfRepository.EXPECT().PathInUseByUser("user-1", "profile", "").Return(true, nil)

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("profile"))

		require.ErrorIs(t, err, ErrConflict)
	})

	t.Run("route words are fine behind a username", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1", UserBase: model.UserBase{Username: "alice"}}, nil)
		svc.ShelfRepository.EXPECT().PathInUseByUser("user-1", "docs", "").Return(false, nil)
		svc.ShelfRepository.EXPECT().Create(gomock.Any()).Return("shelf-1", nil)

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("docs"))

		require.NoError(t, err)
	})

	t.Run("an owner without a username cannot create a shelf with a path", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("profile"))

		require.ErrorIs(t, err, ErrInvalidInput)
		require.ErrorContains(t, err, "username")
	})

	t.Run("a domain-only shelf needs no username", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
		svc.ShelfRepository.EXPECT().DomainInUse("profile.example.com", "").Return(false, nil)
		svc.ShelfRepository.EXPECT().Create(gomock.Any()).Return("shelf-1", nil)

		_, err := svc.Service.ShelfService.Create("user-1", shelfWithDomain("profile.example.com"))

		require.NoError(t, err)
	})
}

// expectShelfUpdateLookups sets up the reads Update performs before it
// validates anything: the caller, and the stored shelf owned by user-1.
func expectShelfUpdateLookups(svc *MockService, storedPath string) {
	svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
	svc.ShelfRepository.EXPECT().Get("shelf-1").Return(&model.Shelf{
		PublicShelf: model.PublicShelf{Id: "shelf-1", Path: storedPath},
		UserId:      "user-1",
	}, nil)
}

func Test_Unit_Shelf_Update_PathRules(t *testing.T) {
	t.Run("a changed path is checked against the whole instance while the feature is off", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectShelfUpdateLookups(svc, "old-path")
		svc.ShelfRepository.EXPECT().PathInUse("new-path", "shelf-1").Return(true, nil)

		shelf, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithPath("new-path"))

		require.ErrorIs(t, err, ErrConflict)
		require.Nil(t, shelf)
	})

	t.Run("a changed path may not become a route word", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectShelfUpdateLookups(svc, "old-path")

		shelf, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithPath("docs"))

		require.ErrorIs(t, err, ErrInvalidInput)
		require.Nil(t, shelf)
	})

	t.Run("a shelf that already has a now-reserved path can still be edited", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		expectShelfUpdateLookups(svc, "docs")
		svc.ShelfRepository.EXPECT().Update(gomock.Any()).Return(nil)
		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-1", Path: "docs", Title: "shelf-title-test"},
			UserId:      "user-1",
		}, nil)

		// No PathInUse expectation: the unchanged path is not re-checked.
		shelf, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithPath("docs"))

		require.NoError(t, err)
		require.NotNil(t, shelf)
	})

	t.Run("with user-based paths only the owner's other shelves count", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		expectShelfUpdateLookups(svc, "old-path")
		svc.ShelfRepository.EXPECT().PathInUseByUser("user-1", "new-path", "shelf-1").Return(true, nil)

		_, err := svc.Service.ShelfService.Update("shelf-1", "user-1", false, shelfWithPath("new-path"))

		require.ErrorIs(t, err, ErrConflict)
	})

	t.Run("an admin editing someone else's shelf is checked against the owner", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.UserRepository.EXPECT().Get("admin-1").Return(&model.User{Id: "admin-1"}, nil)
		svc.ShelfRepository.EXPECT().Get("shelf-1").Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-1", Path: "old-path"},
			UserId:      "user-1",
		}, nil)
		svc.ShelfRepository.EXPECT().PathInUseByUser("user-1", "new-path", "shelf-1").Return(true, nil)

		_, err := svc.Service.ShelfService.Update("shelf-1", "admin-1", true, shelfWithPath("new-path"))

		require.ErrorIs(t, err, ErrConflict)
	})
}

func Test_Unit_Shelf_GetByPath_DoesNotResolveWhileUserBasedPathsIsOn(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	userBasedPaths(t, true)

	// No repository expectation: /<path> must not even be looked up.
	shelf, err := svc.Service.ShelfService.GetByPath("my-path")

	require.NoError(t, err)
	require.Nil(t, shelf)
}

func Test_Unit_Shelf_GetByUsernameAndPath(t *testing.T) {
	t.Run("does not resolve while the feature is off", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, false)

		shelf, err := svc.Service.ShelfService.GetByUsernameAndPath("alice", "my-path")

		require.NoError(t, err)
		require.Nil(t, shelf)
	})

	t.Run("resolves with a lowercased username and resolves the theme", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.ShelfRepository.EXPECT().GetByUsernameAndPath("alice", "my-path").Return(&model.Shelf{
			PublicShelf: model.PublicShelf{Id: "shelf-1", Path: "my-path"},
			ThemeId:     "theme-1",
			Username:    "alice",
		}, nil)
		svc.ThemeRepository.EXPECT().Get("theme-1").Return(&model.Theme{Id: "theme-1", Config: "--shelf-bg: #1c274c;\n"}, nil)

		shelf, err := svc.Service.ShelfService.GetByUsernameAndPath("Alice", "my-path")

		require.NoError(t, err)
		require.NotNil(t, shelf)
		require.Equal(t, "#1c274c", shelf.Theme["--shelf-bg"])
	})

	t.Run("not found", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.ShelfRepository.EXPECT().GetByUsernameAndPath("alice", "missing").Return(nil, nil)

		shelf, err := svc.Service.ShelfService.GetByUsernameAndPath("alice", "missing")

		require.NoError(t, err)
		require.Nil(t, shelf)
	})

	t.Run("repository failure", func(t *testing.T) {
		svc := NewMockService(t)
		defer svc.Ctrl.Finish()
		userBasedPaths(t, true)

		svc.ShelfRepository.EXPECT().GetByUsernameAndPath("alice", "my-path").Return(nil, errors.New("db unavailable"))

		shelf, err := svc.Service.ShelfService.GetByUsernameAndPath("alice", "my-path")

		require.ErrorContains(t, err, "db unavailable")
		require.Nil(t, shelf)
	})
}

func Test_Unit_Shelf_Creation_RecordsTheModeItWasCreatedIn(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "user-based paths off", true: "user-based paths on"}[enabled], func(t *testing.T) {
			svc := NewMockService(t)
			defer svc.Ctrl.Finish()
			userBasedPaths(t, enabled)

			svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1", UserBase: model.UserBase{Username: "alice"}}, nil)
			if enabled {
				svc.ShelfRepository.EXPECT().PathInUseByUser("user-1", "my-path", "").Return(false, nil)
			} else {
				svc.ShelfRepository.EXPECT().PathInUse("my-path", "").Return(false, nil)
			}
			svc.ShelfRepository.EXPECT().Create(gomock.Any()).DoAndReturn(func(s *model.Shelf) (string, error) {
				require.Equal(t, enabled, s.CreatedWithUserBasedPaths)
				return "shelf-1", nil
			})

			_, err := svc.Service.ShelfService.Create("user-1", shelfWithPath("my-path"))

			require.NoError(t, err)
		})
	}
}

func Test_Unit_Shelf_Creation_IgnoresAModeSuppliedByTheClient(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()
	userBasedPaths(t, false)

	svc.UserRepository.EXPECT().Get("user-1").Return(&model.User{Id: "user-1"}, nil)
	svc.ShelfRepository.EXPECT().PathInUse("my-path", "").Return(false, nil)
	svc.ShelfRepository.EXPECT().Create(gomock.Any()).DoAndReturn(func(s *model.Shelf) (string, error) {
		require.False(t, s.CreatedWithUserBasedPaths, "the instance decides, not the request")
		return "shelf-1", nil
	})

	request := shelfWithPath("my-path")
	request.CreatedWithUserBasedPaths = true
	_, err := svc.Service.ShelfService.Create("user-1", request)

	require.NoError(t, err)
}
