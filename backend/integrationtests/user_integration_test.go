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
	request := model.UserCreate{
		UserBase: model.UserBase{
			Email:     "user-api-creation@test.com",
			FirstName: "user-api-creation-firstname",
			LastName:  "user-api-creation-lastname",
		},
		Password: "secret",
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
	require.Equal(t, model.RoleUser, userResp.Role)
}

func Test_API_User_List(t *testing.T) {
	user := &model.UserCreate{
		UserBase: model.UserBase{
			Email:     "user-api-list@test.com",
			FirstName: "user-api-list-firstname",
			LastName:  "user-api-list-lastname",
		},
		Password: "secret",
	}

	created, err := TestService.UserService.Create(user, false)
	require.NoError(t, err)

	_, adminToken := createTestAdmin(t)

	resp := doAuthedRequest(
		t,
		http.MethodGet,
		"/v1/users",
		nil,
		adminToken,
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

	var usersResp []model.User
	err = json.Unmarshal(body, &usersResp)
	require.NoError(t, err)

	found := false
	for _, u := range usersResp {
		if u.Id == created.Id {
			found = true
		}
	}
	require.True(t, found)
}

func Test_API_User_Get(t *testing.T) {
	user := &model.UserCreate{
		UserBase: model.UserBase{
			Email:     "user-api-get@test.com",
			FirstName: "user-api-get-firstname",
			LastName:  "user-api-get-lastname",
		},
		Password: "secret",
	}

	created, err := TestService.UserService.Create(user, false)
	require.NoError(t, err)

	_, adminToken := createTestAdmin(t)

	resp := doAuthedRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/v1/users/%s", created.Id),
		nil,
		adminToken,
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

	require.Equal(t, created.Email, userResp.Email)
	require.Equal(t, created.FirstName, userResp.FirstName)
	require.Equal(t, created.LastName, userResp.LastName)
}

func Test_API_User_Update(t *testing.T) {
	user := &model.UserCreate{
		UserBase: model.UserBase{
			Email:     "user-api-update@test.com",
			FirstName: "user-api-update-firstname",
			LastName:  "user-api-update-lastname",
		},
		Password: "secret",
	}

	created, err := TestService.UserService.Create(user, false)
	require.NoError(t, err)

	// Only the account owner (or an admin) may update a profile - log in as
	// the user itself to exercise the realistic self-update path.
	token := loginAndGetToken(t, user.Email, user.Password)

	updateRequest := model.UserBase{
		Email:     "user-api-update@test.com",
		FirstName: "user-api-update-firstname-updated",
		LastName:  "user-api-update-lastname-updated",
	}

	resp := doAuthedRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/v1/users/%s", created.Id),
		strings.NewReader(ObjectToJSON(updateRequest)),
		token,
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
}

func Test_API_User_PatchPassword(t *testing.T) {
	user := &model.UserCreate{
		UserBase: model.UserBase{
			Email:     "user-api-patch-password@test.com",
			FirstName: "user-api-patch-password-firstname",
			LastName:  "user-api-patch-password-lastname",
		},
		Password: "secret",
	}

	created, err := TestService.UserService.Create(user, false)
	require.NoError(t, err)

	// Only the account owner may patch their own password.
	token := loginAndGetToken(t, user.Email, user.Password)

	patchPasswordRequest := model.UserRequestBodyOnlyPassword{
		OldPassword: "secret",
		NewPassword: "newSecret",
	}

	resp := doAuthedRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/v1/users/%s/password", created.Id),
		strings.NewReader(ObjectToJSON(patchPasswordRequest)),
		token,
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
	user := &model.UserCreate{
		UserBase: model.UserBase{
			Email:     "user-api-delete-user@test.com",
			FirstName: "user-api-delete-user-firstname",
			LastName:  "user-api-delete-user-lastname",
		},
		Password: "secret",
	}

	created, err := TestService.UserService.Create(user, false)
	require.NoError(t, err)

	_, adminToken := createTestAdmin(t)

	resp := doAuthedRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/v1/users/%s", created.Id),
		nil,
		adminToken,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

}
