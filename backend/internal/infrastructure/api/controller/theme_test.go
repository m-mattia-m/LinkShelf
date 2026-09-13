package controller

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_API_ListThemes_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := ListThemes(svc.Service)

	svc.ThemeService.
		EXPECT().
		ListGrouped(gomock.Any()).
		Return(model.ThemeGroupedResponseBody{
			Instance: []model.Theme{{Id: "instance-1", Name: "Midnight"}},
			Mine:     []model.Theme{{Id: "mine-1", Name: "Sunset"}},
		}, nil)

	resp, err := handler(context.Background(), &struct{}{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Body.Instance, 1)
	require.Len(t, resp.Body.Mine, 1)
}

func Test_API_ListThemes_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := ListThemes(svc.Service)

	svc.ThemeService.
		EXPECT().
		ListGrouped(gomock.Any()).
		Return(model.ThemeGroupedResponseBody{}, errors.New("db unavailable"))

	resp, err := handler(context.Background(), &struct{}{})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to list themes")
}

func Test_API_ListAllUserThemes_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := ListAllUserThemes(svc.Service)

	svc.ThemeService.
		EXPECT().
		ListAllUserScoped().
		Return([]model.Theme{{Id: "theme-1"}, {Id: "theme-2"}}, nil)

	resp, err := handler(context.Background(), &struct{}{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Body, 2)
}

func Test_API_ListAllUserThemes_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := ListAllUserThemes(svc.Service)

	svc.ThemeService.
		EXPECT().
		ListAllUserScoped().
		Return(nil, errors.New("db unavailable"))

	resp, err := handler(context.Background(), &struct{}{})

	require.Error(t, err)
	require.Nil(t, resp)
}

func Test_API_GetTheme_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetTheme(svc.Service)

	svc.ThemeService.
		EXPECT().
		Get("theme-uuid-test", gomock.Any(), gomock.Any()).
		Return(&model.Theme{Id: "theme-uuid-test", Name: "Sunset"}, nil)

	resp, err := handler(context.Background(), &model.ThemeRequestFilter{ThemeId: "theme-uuid-test"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "Sunset", resp.Body.Name)
}

func Test_API_GetTheme_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetTheme(svc.Service)

	svc.ThemeService.
		EXPECT().
		Get("theme-uuid-test", gomock.Any(), gomock.Any()).
		Return(nil, errors.New("failed to get theme"))

	resp, err := handler(context.Background(), &model.ThemeRequestFilter{ThemeId: "theme-uuid-test"})

	require.Error(t, err)
	require.Nil(t, resp)
}

func Test_API_CreateTheme_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateTheme(svc.Service)

	input := &model.ThemeRequestBody{
		Body: model.ThemeBase{Name: "Sunset", Config: "--shelf-bg: #1c274c;"},
	}

	svc.ThemeService.
		EXPECT().
		Create(gomock.Any(), input.Body).
		Return(&model.Theme{Id: "theme-uuid-test", Name: "Sunset"}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "theme-uuid-test", resp.Body.Id)
}

func Test_API_CreateTheme_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := CreateTheme(svc.Service)

	input := &model.ThemeRequestBody{
		Body: model.ThemeBase{Name: "Sunset", Config: "--shelf-bg: #1c274c;"},
	}

	svc.ThemeService.
		EXPECT().
		Create(gomock.Any(), input.Body).
		Return(nil, errors.New("invalid input"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
}

func Test_API_UpdateTheme_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateTheme(svc.Service)

	input := &model.ThemeFilterAndBody{
		ThemeRequestFilter: model.ThemeRequestFilter{ThemeId: "theme-uuid-test"},
		Body:               model.ThemeBase{Name: "Updated", Config: "--shelf-bg: #fff;"},
	}

	svc.ThemeService.
		EXPECT().
		Update("theme-uuid-test", gomock.Any(), input.Body).
		Return(&model.Theme{Id: "theme-uuid-test", Name: "Updated"}, nil)

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "Updated", resp.Body.Name)
}

func Test_API_UpdateTheme_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateTheme(svc.Service)

	input := &model.ThemeFilterAndBody{
		ThemeRequestFilter: model.ThemeRequestFilter{ThemeId: "theme-uuid-test"},
		Body:               model.ThemeBase{Name: "Updated", Config: "--shelf-bg: #fff;"},
	}

	svc.ThemeService.
		EXPECT().
		Update("theme-uuid-test", gomock.Any(), input.Body).
		Return(nil, errors.New("failed to update theme"))

	resp, err := handler(context.Background(), input)

	require.Error(t, err)
	require.Nil(t, resp)
}

func Test_API_DeleteTheme_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteTheme(svc.Service)

	svc.ThemeService.
		EXPECT().
		Delete("theme-uuid-test", gomock.Any(), gomock.Any()).
		Return(nil)

	resp, err := handler(context.Background(), &model.ThemeRequestFilter{ThemeId: "theme-uuid-test"})

	require.NoError(t, err)
	require.Nil(t, resp)
}

func Test_API_DeleteTheme_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := DeleteTheme(svc.Service)

	svc.ThemeService.
		EXPECT().
		Delete("theme-uuid-test", gomock.Any(), gomock.Any()).
		Return(errors.New("failed to delete theme"))

	resp, err := handler(context.Background(), &model.ThemeRequestFilter{ThemeId: "theme-uuid-test"})

	require.Error(t, err)
	require.Nil(t, resp)
}
