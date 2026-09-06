package mapper

import (
	"backend/internal/infrastructure/api/model"
)

func MapShelfBaseToShelfPointer(base model.ShelfBase) *model.Shelf {
	return &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       base.Title,
			Description: base.Description,
			Icon:        base.Icon,
			Path:        base.Path,
		},
		Domain: base.Domain,
		Theme:  base.Theme,
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

func MapShelfToPublicShelfResponse(shelf model.Shelf) *model.PublicShelfResponse {
	return &model.PublicShelfResponse{
		Body: shelf.PublicShelf,
	}
}
