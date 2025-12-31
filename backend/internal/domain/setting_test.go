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

	svc.ShelfRepository.
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
		Key:   "theme",
		Value: "dark",
	}

	svc.SettingRepository.
		EXPECT().
		Update("theme", "dark").
		Return(nil)

	err := svc.Service.SettingService.Update(setting)

	require.NoError(t, err)
}

func Test_Unit_Setting_Update_Failure(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	setting := model.Setting{
		Key:   "theme",
		Value: "dark",
	}

	svc.SettingRepository.
		EXPECT().
		Update("theme", "dark").
		Return(errors.New("an error occurred"))

	err := svc.Service.SettingService.Update(setting)

	require.ErrorContains(t, err, "an error occurred")
}
