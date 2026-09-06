package mapper

import "backend/internal/infrastructure/api/model"

func MapStatisticToStatisticResponse(body model.Statistic) *model.StatisticResponse {
	return &model.StatisticResponse{
		Body: body,
	}
}
