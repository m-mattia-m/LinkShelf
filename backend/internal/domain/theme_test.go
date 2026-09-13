package domain

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Unit_Theme_ListGrouped_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		ListInstance().
		Return([]model.Theme{{Id: "instance-1", Scope: model.ThemeScopeInstance, Name: "Midnight"}}, nil)

	svc.ThemeRepository.
		EXPECT().
		ListByOwner("user-uuid-test").
		Return([]model.Theme{{Id: "mine-1", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test", Name: "Sunset"}}, nil)

	grouped, err := svc.Service.ThemeService.ListGrouped("user-uuid-test")

	require.NoError(t, err)
	require.Len(t, grouped.Instance, 1)
	require.Len(t, grouped.Mine, 1)
	require.Equal(t, "Midnight", grouped.Instance[0].Name)
	require.Equal(t, "Sunset", grouped.Mine[0].Name)
}

func Test_Unit_Theme_ListGrouped_InstanceListFailure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		ListInstance().
		Return(nil, errors.New("db unavailable"))

	_, err := svc.Service.ThemeService.ListGrouped("user-uuid-test")

	require.ErrorContains(t, err, "db unavailable")
}

func Test_Unit_Theme_ListGrouped_OwnerListFailure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		ListInstance().
		Return([]model.Theme{}, nil)

	svc.ThemeRepository.
		EXPECT().
		ListByOwner("user-uuid-test").
		Return(nil, errors.New("db unavailable"))

	_, err := svc.Service.ThemeService.ListGrouped("user-uuid-test")

	require.ErrorContains(t, err, "db unavailable")
}

func Test_Unit_Theme_Get_InstanceScope_AnyoneCanRead(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeInstance, Name: "Midnight"}, nil)

	theme, err := svc.Service.ThemeService.Get("theme-uuid-test", "someone-uuid-test", false)

	require.NoError(t, err)
	require.NotNil(t, theme)
	require.Equal(t, "Midnight", theme.Name)
}

func Test_Unit_Theme_Get_UserScope_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test", Name: "Sunset"}, nil)

	theme, err := svc.Service.ThemeService.Get("theme-uuid-test", "user-uuid-test", false)

	require.NoError(t, err)
	require.NotNil(t, theme)
}

func Test_Unit_Theme_Get_UserScope_Admin(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "owner-uuid-test"}, nil)

	theme, err := svc.Service.ThemeService.Get("theme-uuid-test", "admin-uuid-test", true)

	require.NoError(t, err)
	require.NotNil(t, theme)
}

func Test_Unit_Theme_Get_UserScope_Forbidden(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "owner-uuid-test"}, nil)

	theme, err := svc.Service.ThemeService.Get("theme-uuid-test", "someone-else-uuid-test", false)

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, theme)
}

func Test_Unit_Theme_Get_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, nil)

	theme, err := svc.Service.ThemeService.Get("theme-uuid-test", "user-uuid-test", false)

	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, theme)
}

func Test_Unit_Theme_Get_RepositoryFailure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, errors.New("db unavailable"))

	theme, err := svc.Service.ThemeService.Get("theme-uuid-test", "user-uuid-test", false)

	require.ErrorContains(t, err, "db unavailable")
	require.Nil(t, theme)
}

func Test_Unit_Theme_Create_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Create(&model.Theme{
			Scope:       model.ThemeScopeUser,
			OwnerUserId: "user-uuid-test",
			Name:        "Sunset",
			Config:      "--shelf-bg: #1c274c;\n",
		}).
		Return("theme-uuid-test", nil)

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test", Name: "Sunset", Config: "--shelf-bg: #1c274c;\n"}, nil)

	theme, err := svc.Service.ThemeService.Create("user-uuid-test", model.ThemeBase{Name: "Sunset", Config: "--shelf-bg: #1c274c;"})

	require.NoError(t, err)
	require.NotNil(t, theme)
	require.Equal(t, "Sunset", theme.Name)
}

func Test_Unit_Theme_Create_InvalidConfig(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	theme, err := svc.Service.ThemeService.Create("user-uuid-test", model.ThemeBase{Name: "Bad", Config: "--shelf-bg: not-a-color;"})

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, theme)
}

func Test_Unit_Theme_Create_RepositoryFailure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Create(gomock.Any()).
		Return("", errors.New("db unavailable"))

	theme, err := svc.Service.ThemeService.Create("user-uuid-test", model.ThemeBase{Name: "Sunset", Config: "--shelf-bg: #1c274c;"})

	require.ErrorContains(t, err, "db unavailable")
	require.Nil(t, theme)
}

func Test_Unit_Theme_Update_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test", Name: "Old"}, nil)

	svc.ThemeRepository.
		EXPECT().
		Update(gomock.Any()).
		Return(nil)

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test", Name: "New", Config: "--shelf-bg: #fff;\n"}, nil)

	theme, err := svc.Service.ThemeService.Update("theme-uuid-test", "user-uuid-test", model.ThemeBase{Name: "New", Config: "--shelf-bg: #fff;"})

	require.NoError(t, err)
	require.NotNil(t, theme)
	require.Equal(t, "New", theme.Name)
}

func Test_Unit_Theme_Update_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "owner-uuid-test"}, nil)

	theme, err := svc.Service.ThemeService.Update("theme-uuid-test", "someone-else-uuid-test", model.ThemeBase{Name: "New", Config: "--shelf-bg: #fff;"})

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, theme)
}

func Test_Unit_Theme_Update_Forbidden_AdminCannotEditOthersTheme(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "owner-uuid-test"}, nil)

	// Update takes no isAdmin parameter at all - not even an admin may edit
	// someone else's theme content, only delete it (see the Delete tests).
	theme, err := svc.Service.ThemeService.Update("theme-uuid-test", "admin-uuid-test", model.ThemeBase{Name: "New", Config: "--shelf-bg: #fff;"})

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, theme)
}

func Test_Unit_Theme_Update_Forbidden_InstanceScope(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeInstance}, nil)

	theme, err := svc.Service.ThemeService.Update("theme-uuid-test", "user-uuid-test", model.ThemeBase{Name: "New", Config: "--shelf-bg: #fff;"})

	require.ErrorIs(t, err, ErrForbidden)
	require.Nil(t, theme)
}

func Test_Unit_Theme_Update_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, nil)

	theme, err := svc.Service.ThemeService.Update("theme-uuid-test", "user-uuid-test", model.ThemeBase{Name: "New", Config: "--shelf-bg: #fff;"})

	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, theme)
}

func Test_Unit_Theme_Update_InvalidConfig(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test"}, nil)

	theme, err := svc.Service.ThemeService.Update("theme-uuid-test", "user-uuid-test", model.ThemeBase{Name: "New", Config: "not-a-declaration"})

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, theme)
}

func Test_Unit_Theme_Delete_Success_Owner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test"}, nil)

	svc.ThemeRepository.
		EXPECT().
		Delete("theme-uuid-test").
		Return(nil)

	err := svc.Service.ThemeService.Delete("theme-uuid-test", "user-uuid-test", false)

	require.NoError(t, err)
}

func Test_Unit_Theme_Delete_Success_Admin(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "owner-uuid-test"}, nil)

	svc.ThemeRepository.
		EXPECT().
		Delete("theme-uuid-test").
		Return(nil)

	err := svc.Service.ThemeService.Delete("theme-uuid-test", "admin-uuid-test", true)

	require.NoError(t, err)
}

func Test_Unit_Theme_Delete_Forbidden_NotOwner(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "owner-uuid-test"}, nil)

	err := svc.Service.ThemeService.Delete("theme-uuid-test", "someone-else-uuid-test", false)

	require.ErrorIs(t, err, ErrForbidden)
}

func Test_Unit_Theme_Delete_Forbidden_InstanceScope_EvenAsAdmin(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeInstance}, nil)

	err := svc.Service.ThemeService.Delete("theme-uuid-test", "admin-uuid-test", true)

	require.ErrorIs(t, err, ErrForbidden)
}

func Test_Unit_Theme_Delete_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, nil)

	err := svc.Service.ThemeService.Delete("theme-uuid-test", "user-uuid-test", false)

	require.ErrorIs(t, err, ErrNotFound)
}

func Test_Unit_Theme_ListAllUserScoped_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		ListAllUserScoped().
		Return([]model.Theme{{Id: "theme-1"}, {Id: "theme-2"}}, nil)

	themes, err := svc.Service.ThemeService.ListAllUserScoped()

	require.NoError(t, err)
	require.Len(t, themes, 2)
}

func Test_Unit_Theme_ListAllUserScoped_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		ListAllUserScoped().
		Return(nil, errors.New("db unavailable"))

	themes, err := svc.Service.ThemeService.ListAllUserScoped()

	require.ErrorContains(t, err, "db unavailable")
	require.Nil(t, themes)
}

func Test_Unit_Theme_Resolve_EmptyId(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	config, missing, err := svc.Service.ThemeService.Resolve("")

	require.NoError(t, err)
	require.False(t, missing)
	require.Nil(t, config)
}

func Test_Unit_Theme_Resolve_Missing(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, nil)

	config, missing, err := svc.Service.ThemeService.Resolve("theme-uuid-test")

	require.NoError(t, err)
	require.True(t, missing)
	require.Nil(t, config)
}

func Test_Unit_Theme_Resolve_Found(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Config: "--shelf-bg: #1c274c;\n--shelf-text: #ffffff;\n"}, nil)

	config, missing, err := svc.Service.ThemeService.Resolve("theme-uuid-test")

	require.NoError(t, err)
	require.False(t, missing)
	require.Equal(t, "#1c274c", config["--shelf-bg"])
	require.Equal(t, "#ffffff", config["--shelf-text"])
}

func Test_Unit_Theme_Resolve_RepositoryFailure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, errors.New("db unavailable"))

	config, missing, err := svc.Service.ThemeService.Resolve("theme-uuid-test")

	require.ErrorContains(t, err, "db unavailable")
	require.False(t, missing)
	require.Nil(t, config)
}

func Test_Unit_Theme_ValidateAssignable_EmptyId(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	require.NoError(t, svc.Service.ThemeService.ValidateAssignable("", "user-uuid-test"))
}

func Test_Unit_Theme_ValidateAssignable_NotFound(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, nil)

	err := svc.Service.ThemeService.ValidateAssignable("theme-uuid-test", "user-uuid-test")

	require.ErrorIs(t, err, ErrInvalidInput)
}

func Test_Unit_Theme_ValidateAssignable_InstanceScope_Ok(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeInstance}, nil)

	err := svc.Service.ThemeService.ValidateAssignable("theme-uuid-test", "user-uuid-test")

	require.NoError(t, err)
}

func Test_Unit_Theme_ValidateAssignable_OwnedByShelfOwner_Ok(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "user-uuid-test"}, nil)

	err := svc.Service.ThemeService.ValidateAssignable("theme-uuid-test", "user-uuid-test")

	require.NoError(t, err)
}

func Test_Unit_Theme_ValidateAssignable_NotOwnedByShelfOwner_Forbidden(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(&model.Theme{Id: "theme-uuid-test", Scope: model.ThemeScopeUser, OwnerUserId: "someone-else-uuid-test"}, nil)

	err := svc.Service.ThemeService.ValidateAssignable("theme-uuid-test", "user-uuid-test")

	require.ErrorIs(t, err, ErrForbidden)
}

func Test_Unit_Theme_ValidateAssignable_RepositoryFailure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.ThemeRepository.
		EXPECT().
		Get("theme-uuid-test").
		Return(nil, errors.New("db unavailable"))

	err := svc.Service.ThemeService.ValidateAssignable("theme-uuid-test", "user-uuid-test")

	require.ErrorContains(t, err, "db unavailable")
}
