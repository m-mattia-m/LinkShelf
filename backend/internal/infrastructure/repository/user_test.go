package repository

import (
	"backend/internal/infrastructure/api/model"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	if testRepo == nil {
		t.Fatal("repository not initialized")
	}

	userId, err := testRepo.UserRepository.Create(&model.User{
		Id: uuid.New().String(),
		UserBase: model.UserBase{
			Email:     "user@test.com",
			FirstName: "John",
			LastName:  "Doe",
			Password:  "userpassword",
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, userId)

}
