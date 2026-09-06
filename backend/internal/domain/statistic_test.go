package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr(i int) *int {
	return &i
}

func Test_Unit_Statistic_Get_Success(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.StatisticRepository.
		EXPECT().
		GetShelfAmount("user-uuid-test").
		Return(ptr(3), nil)

	svc.StatisticRepository.
		EXPECT().
		GetSectionAmount("user-uuid-test").
		Return(ptr(5), nil)

	svc.StatisticRepository.
		EXPECT().
		GetLinkAmount("user-uuid-test").
		Return(ptr(8), nil)

	statistic, err := svc.Service.StatisticService.Get("user-uuid-test")

	require.NoError(t, err)
	require.NotNil(t, statistic)
	require.Equal(t, 3, statistic.ShelfNumber)
	require.Equal(t, 5, statistic.SectionNumber)
	require.Equal(t, 8, statistic.LinkNumber)
}

func Test_Unit_Statistic_Get_Failure_ShelfAmount(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.StatisticRepository.
		EXPECT().
		GetShelfAmount("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	statistic, err := svc.Service.StatisticService.Get("user-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, statistic)
}

func Test_Unit_Statistic_Get_Failure_SectionAmount(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.StatisticRepository.
		EXPECT().
		GetShelfAmount("user-uuid-test").
		Return(ptr(3), nil)

	svc.StatisticRepository.
		EXPECT().
		GetSectionAmount("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	statistic, err := svc.Service.StatisticService.Get("user-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, statistic)
}

func Test_Unit_Statistic_Get_Failure_LinkAmount(t *testing.T) {
	svc := NewMockService(t)
	defer svc.Ctrl.Finish()

	svc.StatisticRepository.
		EXPECT().
		GetShelfAmount("user-uuid-test").
		Return(ptr(3), nil)

	svc.StatisticRepository.
		EXPECT().
		GetSectionAmount("user-uuid-test").
		Return(ptr(5), nil)

	svc.StatisticRepository.
		EXPECT().
		GetLinkAmount("user-uuid-test").
		Return(nil, errors.New("an error occurred"))

	statistic, err := svc.Service.StatisticService.Get("user-uuid-test")

	require.ErrorContains(t, err, "an error occurred")
	require.Nil(t, statistic)
}
