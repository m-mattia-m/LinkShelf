package mapper

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MapUserBaseToUserPointer(t *testing.T) {
	base := model.UserBase{
		Email:     "test@test.com",
		FirstName: "firstname-test",
		LastName:  "lastname-test",
		Password:  "secret",
	}

	result := MapUserBaseToUserPointer(base)

	require.NotNil(t, result)
	require.Equal(t, base.Email, result.Email)
	require.Equal(t, base.FirstName, result.FirstName)
	require.Equal(t, base.LastName, result.LastName)
	require.Equal(t, base.Password, result.Password)
}

func Test_MapUserToUserResponse(t *testing.T) {
	user := model.User{
		Id: "user-uuid-test",
		UserBase: model.UserBase{
			Email:     "test@test.com",
			FirstName: "firstname-test",
			LastName:  "lastname-test",
		},
	}

	resp := MapUserToUserResponse(user)

	require.NotNil(t, resp)
	require.Equal(t, user.Id, resp.Body.Id)
	require.Equal(t, user.Email, resp.Body.Email)
	require.Equal(t, user.FirstName, resp.Body.FirstName)
	require.Equal(t, user.LastName, resp.Body.LastName)
}
