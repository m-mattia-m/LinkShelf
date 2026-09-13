package mapper

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MapThemeToThemeResponse(t *testing.T) {
	theme := model.Theme{
		Id:     "theme-uuid-test",
		Scope:  model.ThemeScopeUser,
		Name:   "Sunset",
		Config: "--shelf-bg: #1c274c;\n",
	}

	resp := MapThemeToThemeResponse(theme)

	require.NotNil(t, resp)
	require.Equal(t, theme.Id, resp.Body.Id)
	require.Equal(t, theme.Name, resp.Body.Name)
	require.Equal(t, theme.Config, resp.Body.Config)
}

func Test_MapThemesToThemeListResponse(t *testing.T) {
	themes := []model.Theme{
		{Id: "theme-1", Name: "First"},
		{Id: "theme-2", Name: "Second"},
	}

	resp := MapThemesToThemeListResponse(themes)

	require.NotNil(t, resp)
	require.Len(t, resp.Body, 2)
	require.Equal(t, "theme-1", resp.Body[0].Id)
	require.Equal(t, "theme-2", resp.Body[1].Id)
}

func Test_MapThemesToThemeListResponse_Empty(t *testing.T) {
	resp := MapThemesToThemeListResponse([]model.Theme{})

	require.NotNil(t, resp)
	require.Empty(t, resp.Body)
}

func Test_MapThemeGroupedToResponse(t *testing.T) {
	body := model.ThemeGroupedResponseBody{
		Instance: []model.Theme{{Id: "instance-1", Name: "Midnight"}},
		Mine:     []model.Theme{{Id: "mine-1", Name: "Sunset"}},
	}

	resp := MapThemeGroupedToResponse(body)

	require.NotNil(t, resp)
	require.Len(t, resp.Body.Instance, 1)
	require.Len(t, resp.Body.Mine, 1)
	require.Equal(t, "Midnight", resp.Body.Instance[0].Name)
	require.Equal(t, "Sunset", resp.Body.Mine[0].Name)
}
