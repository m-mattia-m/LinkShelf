package mapper

import (
	"backend/internal/infrastructure/api/model"
)

func MapSectionBaseToSectionPointer(base model.SectionBase) *model.Section {
	return &model.Section{
		SectionBase: base,
	}
}

func MapSectionToSectionResponse(body model.Section) *model.SectionResponse {
	return &model.SectionResponse{
		Body: body,
	}
}

func MapSectionsToSectionResponseList(sections []model.Section) *model.SectionResponseList {
	return &model.SectionResponseList{Body: sections}
}
