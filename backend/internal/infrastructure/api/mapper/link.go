package mapper

import (
	"backend/internal/infrastructure/api/model"
)

func MapLinkBaseToLinkPointer(base model.LinkBase) *model.Link {
	return &model.Link{
		LinkBase: base,
	}
}

func MapLinkToLinkResponse(body model.Link) *model.LinkResponse {
	return &model.LinkResponse{
		Body: body,
	}
}

func MapLinksToLinkResponseList(links []model.Link) *model.LinkResponseList {
	return &model.LinkResponseList{
		Body: links,
	}
}
