package mapper

import (
	"backend/internal/infrastructure/api/model"
)

func MapThemeToThemeResponse(t model.Theme) *model.ThemeResponse {
	return &model.ThemeResponse{Body: t}
}

func MapThemesToThemeListResponse(themes []model.Theme) *model.ThemeListResponse {
	return &model.ThemeListResponse{Body: themes}
}

func MapThemeGroupedToResponse(body model.ThemeGroupedResponseBody) *model.ThemeGroupedResponse {
	return &model.ThemeGroupedResponse{Body: body}
}
