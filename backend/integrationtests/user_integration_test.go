//go:build integration
// +build integration

package integrationtests

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

func Test_API_User_Create(t *testing.T) {
	request := model.UserBase{
		Email:     "user-api-creation@test.com",
		FirstName: "user-api-creation-firstname",
		LastName:  "user-api-creation-lastname",
		Password:  "secret",
	}

	resp := doRequest(
		t,
		http.MethodPost,
		"/v1/users",
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

func Test_API_User_Get(t *testing.T) {
	user := &model.User{
		UserBase: model.UserBase{
			Email:     "user-api-get@test.com",
			FirstName: "user-api-get-firstname",
			LastName:  "user-api-get-lastname",
			Password:  "secret",
		},
	}

	user, err := TestService.UserService.Create(user)
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/v1/users/%s", user.Id),
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

func Test_API_User_Update(t *testing.T) {
	user := &model.User{
		UserBase: model.UserBase{
			Email:     "user-api-update@test.com",
			FirstName: "user-api-update-firstname",
			LastName:  "user-api-update-lastname",
			Password:  "secret",
		},
	}

	user, err := TestService.UserService.Create(user)
	require.NoError(t, err)

	updateRequest := model.UserBase{
		Email:     "",
		FirstName: "user-api-update-firstname-updated",
		LastName:  "user-api-update-lastname-updated",
		Password:  "",
	}

	resp := doRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/v1/users/%s", user.Id),
		strings.NewReader(ObjectToJSON(updateRequest)),
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

	require.Equal(t, updateRequest.FirstName, userResp.FirstName)
	require.Equal(t, updateRequest.LastName, userResp.LastName)
	require.Empty(t, userResp.Password)

}

func Test_API_User_PatchPassword(t *testing.T) {
	user := &model.User{
		UserBase: model.UserBase{
			Email:     "user-api-patch-password@test.com",
			FirstName: "user-api-patch-password-firstname",
			LastName:  "user-api-patch-password-lastname",
			Password:  "secret",
		},
	}

	user, err := TestService.UserService.Create(user)
	require.NoError(t, err)

	patchPasswordRequest := model.UserRequestBodyOnlyPassword{
		OldPassword: "secret",
		NewPassword: "newSecret",
	}

	resp := doRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/v1/users/%s/password", user.Id),
		strings.NewReader(ObjectToJSON(patchPasswordRequest)),
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyString := string(body)

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	_ = bodyString

}

func Test_API_User_Delete(t *testing.T) {
	user := &model.User{
		UserBase: model.UserBase{
			Email:     "user-api-delete-user@test.com",
			FirstName: "user-api-delete-user-firstname",
			LastName:  "user-api-delete-user-lastname",
			Password:  "secret",
		},
	}

	user, err := TestService.UserService.Create(user)
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/v1/users/%s", user.Id),
		nil,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

}
