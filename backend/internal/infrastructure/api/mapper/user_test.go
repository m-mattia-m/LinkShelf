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
	}

	result := MapUserBaseToUserPointer(base)

	require.NotNil(t, result)
	require.Equal(t, base.Email, result.Email)
	require.Equal(t, base.FirstName, result.FirstName)
	require.Equal(t, base.LastName, result.LastName)
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

func Test_MapUsersToUserListResponse(t *testing.T) {
	users := []model.User{
		{Id: "user-uuid-1", UserBase: model.UserBase{Email: "one@test.com"}},
		{Id: "user-uuid-2", UserBase: model.UserBase{Email: "two@test.com"}},
	}

	resp := MapUsersToUserListResponse(users)

	require.NotNil(t, resp)
	require.Len(t, resp.Body, 2)
	require.Equal(t, "user-uuid-1", resp.Body[0].Id)
	require.Equal(t, "user-uuid-2", resp.Body[1].Id)
}
