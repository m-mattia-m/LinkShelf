package domain

import (
	"backend/internal/infrastructure/api/model"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Unit_Setting_List_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	expected := []model.Setting{
		{Key: "theme", Value: "dark"},
		{Key: "language", Value: "en"},
	}

	svc.SettingRepository.
		EXPECT().
		List().
		Return(expected, nil)

	settings, err := svc.Service.SettingService.List()

	require.NoError(t, err)
	require.Len(t, settings, 2)
	require.Equal(t, "theme", settings[0].Key)
	require.Equal(t, "dark", settings[0].Value)
}

func Test_Unit_Setting_List_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SettingRepository.
		EXPECT().
		List().
		Return(nil, errors.New("an error occurred"))

	settings, err := svc.Service.SettingService.List()

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, settings)
}

func Test_Unit_Setting_Update_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	setting := model.Setting{
		Key:          "theme",
		LanguageCode: "en",
		Value:        "dark",
	}

	svc.SettingRepository.
		EXPECT().
		Upsert("theme", "en", "dark").
		Return(nil)

	err := svc.Service.SettingService.Update(setting)

	require.NoError(t, err)
}

func Test_Unit_Setting_Update_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	setting := model.Setting{
		Key:          "theme",
		LanguageCode: "en",
		Value:        "dark",
	}

	svc.SettingRepository.
		EXPECT().
		Upsert("theme", "en", "dark").
		Return(errors.New("an error occurred"))

	err := svc.Service.SettingService.Update(setting)

	require.ErrorContains(t, err, "an error occurred")
}

func Test_Unit_Setting_UpdateMany_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SettingRepository.
		EXPECT().
		Upsert("about", "en", "Welcome").
		Return(nil)

	svc.SettingRepository.
		EXPECT().
		Upsert("about_show", "en", "true").
		Return(nil)

	failures := svc.Service.SettingService.UpdateMany([]model.Setting{
		{Key: "about", LanguageCode: "en", Value: "Welcome"},
		{Key: "about_show", LanguageCode: "en", Value: "true"},
	})

	require.Empty(t, failures)
}

func Test_Unit_Setting_UpdateMany_UnknownKey(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	failures := svc.Service.SettingService.UpdateMany([]model.Setting{
		{Key: "not_a_real_key", LanguageCode: "en", Value: "anything"},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "not_a_real_key", failures[0].Key)
	require.Contains(t, failures[0].Reason, "unknown setting key")
}

func Test_Unit_Setting_UpdateMany_InvalidBooleanValue(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	failures := svc.Service.SettingService.UpdateMany([]model.Setting{
		{Key: "about_show", LanguageCode: "en", Value: "yes"},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "about_show", failures[0].Key)
	require.Contains(t, failures[0].Reason, `must be "true" or "false"`)
}

func Test_Unit_Setting_UpdateMany_UnsupportedLanguage(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	failures := svc.Service.SettingService.UpdateMany([]model.Setting{
		{Key: "about", LanguageCode: "fr", Value: "Bienvenue"},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "about", failures[0].Key)
	require.Contains(t, failures[0].Reason, "unsupported language code")
}

func Test_Unit_Setting_UpdateMany_UpsertError(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SettingRepository.
		EXPECT().
		Upsert("about", "en", "Welcome").
		Return(errors.New("db unavailable"))

	failures := svc.Service.SettingService.UpdateMany([]model.Setting{
		{Key: "about", LanguageCode: "en", Value: "Welcome"},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "about", failures[0].Key)
	require.Contains(t, failures[0].Reason, "db unavailable")
}

func Test_Unit_Setting_UpdateMany_PartialSuccess(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.SettingRepository.
		EXPECT().
		Upsert("about", "en", "Welcome").
		Return(nil)

	failures := svc.Service.SettingService.UpdateMany([]model.Setting{
		{Key: "about", LanguageCode: "en", Value: "Welcome"},
		{Key: "about_show", LanguageCode: "en", Value: "not-a-bool"},
	})

	require.Len(t, failures, 1)
	require.Equal(t, "about_show", failures[0].Key)
}
