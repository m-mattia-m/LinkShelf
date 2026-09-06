package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"

	"github.com/danielgtaylor/huma/v2"
)

func GetStatistic(svc *domain.Service) func(c context.Context, input *struct{}) (*model.StatisticResponse, error) {
	return func(c context.Context, input *struct{}) (*model.StatisticResponse, error) {
		statistic, err := svc.StatisticService.Get(UserIdFromContext(c))
		if err != nil {
			return nil, huma.Error400BadRequest("failed to get statistic", err)
		}

		return mapper.MapStatisticToStatisticResponse(*statistic), nil
	}
}
