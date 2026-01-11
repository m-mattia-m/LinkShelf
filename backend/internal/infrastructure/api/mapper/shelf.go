package mapper

import (
	"backend/internal/infrastructure/api/model"
)

func MapShelfBaseToShelfPointer(base model.ShelfBase) *model.Shelf {
	return &model.Shelf{
		ShelfBase: base,
	}
}

func MapShelfToShelfResponse(body model.Shelf) *model.ShelfResponse {
	return &model.ShelfResponse{
		Body: body,
	}
}

func MapShelfToShelfListResponse(body []model.Shelf) *model.ShelfListResponse {
	return &model.ShelfListResponse{
		Body: body,
	}
}
