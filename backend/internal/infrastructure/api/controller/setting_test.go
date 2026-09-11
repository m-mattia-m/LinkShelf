package controller

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_API_UpdateSetting_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	gomock.InOrder(
		svc.SettingService.
			EXPECT().
			Update(model.Setting{Key: "about_show", LanguageCode: "en", Value: "true"}).
			Return(nil),

		svc.SettingService.
			EXPECT().
			List().
			Return([]model.Setting{
				{Key: "about_show", LanguageCode: "en", Value: "true"},
			}, nil),
	)
	svc.AuthService.EXPECT().IsOidcEnabled().Return(false)

	handler := UpdateSetting(svc.Service)
	resp, err := handler(context.Background(), &model.SettingRequest{
		Body: model.Setting{Key: "about_show", LanguageCode: "en", Value: "true"},
	})

	require.NoError(t, err)
	require.True(t, resp.Body.AboutShow)
}

func Test_API_UpdateSetting_UpdateFails(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.SettingService.
		EXPECT().
		Update(gomock.Any()).
		Return(errors.New("db unavailable"))

	handler := UpdateSetting(svc.Service)
	_, err := handler(context.Background(), &model.SettingRequest{
		Body: model.Setting{Key: "about_show", LanguageCode: "en", Value: "true"},
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to update setting")
}

func Test_API_UpdateSetting_ListFails(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.SettingService.EXPECT().Update(gomock.Any()).Return(nil)
	svc.SettingService.EXPECT().List().Return(nil, errors.New("db unavailable"))

	handler := UpdateSetting(svc.Service)
	_, err := handler(context.Background(), &model.SettingRequest{
		Body: model.Setting{Key: "about_show", LanguageCode: "en", Value: "true"},
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get setting")
}

func Test_API_UpdateSettingsBatch_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	settings := []model.Setting{
		{Key: "about", LanguageCode: "en", Value: "Welcome"},
		{Key: "about_show", LanguageCode: "en", Value: "true"},
	}

	gomock.InOrder(
		svc.SettingService.
			EXPECT().
			UpdateMany(settings).
			Return(nil),

		svc.SettingService.
			EXPECT().
			List().
			Return(settings, nil),
	)
	svc.AuthService.EXPECT().IsOidcEnabled().Return(false)

	handler := UpdateSettingsBatch(svc.Service)
	resp, err := handler(context.Background(), &model.SettingBatchRequest{
		Body: model.SettingBatchRequestBody{Settings: settings},
	})

	require.NoError(t, err)
	require.Empty(t, resp.Body.Failures)
	require.Equal(t, "Welcome", resp.Body.Settings.About)
	require.True(t, resp.Body.Settings.AboutShow)
}

func Test_API_UpdateSettingsBatch_PartialFailure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	settings := []model.Setting{
		{Key: "about", LanguageCode: "en", Value: "Welcome"},
		{Key: "not_a_real_key", LanguageCode: "en", Value: "x"},
	}
	expectedFailures := []model.SettingUpdateFailure{
		{Key: "not_a_real_key", LanguageCode: "en", Reason: `unknown setting key "not_a_real_key"`},
	}

	gomock.InOrder(
		svc.SettingService.
			EXPECT().
			UpdateMany(settings).
			Return(expectedFailures),

		svc.SettingService.
			EXPECT().
			List().
			Return([]model.Setting{{Key: "about", LanguageCode: "en", Value: "Welcome"}}, nil),
	)
	svc.AuthService.EXPECT().IsOidcEnabled().Return(false)

	handler := UpdateSettingsBatch(svc.Service)
	resp, err := handler(context.Background(), &model.SettingBatchRequest{
		Body: model.SettingBatchRequestBody{Settings: settings},
	})

	require.NoError(t, err)
	require.Equal(t, expectedFailures, resp.Body.Failures)
	require.Equal(t, "Welcome", resp.Body.Settings.About)
}

func Test_API_UpdateSettingsBatch_EmptyBatch(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := UpdateSettingsBatch(svc.Service)
	_, err := handler(context.Background(), &model.SettingBatchRequest{
		Body: model.SettingBatchRequestBody{Settings: []model.Setting{}},
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "settings must not be empty")
}

func Test_API_UpdateSettingsBatch_ListFails(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	settings := []model.Setting{{Key: "about", LanguageCode: "en", Value: "Welcome"}}

	svc.SettingService.EXPECT().UpdateMany(settings).Return(nil)
	svc.SettingService.EXPECT().List().Return(nil, errors.New("db unavailable"))

	handler := UpdateSettingsBatch(svc.Service)
	_, err := handler(context.Background(), &model.SettingBatchRequest{
		Body: model.SettingBatchRequestBody{Settings: settings},
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get settings")
}

func Test_API_GetPageSettings_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.SettingService.
		EXPECT().
		List().
		Return([]model.Setting{
			{Key: "about", LanguageCode: "en", Value: "Welcome"},
		}, nil)
	svc.AuthService.EXPECT().IsOidcEnabled().Return(true)

	handler := GetPageSettings(svc.Service)
	resp, err := handler(context.Background(), &model.SettingRequestFiler{LanguageCode: "en"})

	require.NoError(t, err)
	require.Equal(t, "Welcome", resp.Body.About)
	require.True(t, resp.Body.OidcEnabled)
}

func Test_API_GetPageSettings_ListFails(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.SettingService.EXPECT().List().Return(nil, errors.New("db unavailable"))

	handler := GetPageSettings(svc.Service)
	_, err := handler(context.Background(), &model.SettingRequestFiler{LanguageCode: "en"})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get settings")
}
