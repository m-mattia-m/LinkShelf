package mapper

import (
	"backend/internal/infrastructure/api/model"
)

func MapUserBaseToUserPointer(base model.UserBase) *model.User {
	return &model.User{
		UserBase: base,
	}
}

func MapUserToUserResponse(body model.User) *model.UserResponse {
	return &model.UserResponse{
		Body: body,
	}
}
