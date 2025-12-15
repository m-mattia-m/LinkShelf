package main

import (
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateUser_API(t *testing.T) {
	request := model.UserBase{
		Email:     "user0@test.com",
		FirstName: "user0",
		LastName:  "test",
		Password:  "secret",
	}

	resp := doRequest(
		t,
		http.MethodPost,
		"/v1/user",
		strings.NewReader(ObjectToJSON(request)),
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var userResp model.User
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	require.Equal(t, request.Email, userResp.Email)
	require.Equal(t, request.FirstName, userResp.FirstName)
	require.Equal(t, request.LastName, userResp.LastName)
	require.Empty(t, userResp.Password)
}

func TestGetUser_API(t *testing.T) {
	user := &model.User{
		UserBase: model.UserBase{
			Email:     "user1@test.com",
			FirstName: "user1",
			LastName:  "test",
			Password:  "secret",
		},
	}

	userId, err := TestRepository.UserRepository.Create(user)
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/v1/user/%s", userId),
		strings.NewReader(ObjectToJSON(user.UserBase)),
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var userResp model.User
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	require.Equal(t, user.Email, userResp.Email)
	require.Equal(t, user.FirstName, userResp.FirstName)
	require.Equal(t, user.LastName, userResp.LastName)
	require.Empty(t, userResp.Password)

}
