package controller

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_API_GetStatistic_Success(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetStatistic(svc.Service)

	svc.StatisticService.
		EXPECT().
		Get("TODO").
		Return(&model.Statistic{
			ShelfNumber:   3,
			SectionNumber: 5,
			LinkNumber:    8,
		}, nil)

	resp, err := handler(context.Background(), &struct{}{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, resp.Body.ShelfNumber)
	require.Equal(t, 5, resp.Body.SectionNumber)
	require.Equal(t, 8, resp.Body.LinkNumber)
}

func Test_API_GetStatistic_Failure(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	handler := GetStatistic(svc.Service)

	svc.StatisticService.
		EXPECT().
		Get("TODO").
		Return(nil, errors.New("failed to get statistic"))

	resp, err := handler(context.Background(), &struct{}{})

	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to get statistic")
}
