package mapper

import "backend/internal/infrastructure/api/model"

func MapToHttpSuccess(message string) *model.SuccessResponse {
	return &model.SuccessResponse{
		Body: model.HttpResponseBody{
			Message: message,
		},
	}
}
