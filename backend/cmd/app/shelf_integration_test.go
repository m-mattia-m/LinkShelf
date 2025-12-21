//go:build integration
// +build integration

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

func Test_API_Shelf_Create(t *testing.T) {

	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "shelf-title-creation",
		Path:        "/shelf-title-creation",
		Domain:      "",
		Description: "A shelf created during API integration tests",
		Theme:       "",
		Icon:        "",
		UserId:      userId,
	}

	resp := doRequest(
		t,
		http.MethodPost,
		"/v1/shelves",
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

	var shelfResp model.Shelf
	err = json.Unmarshal(body, &shelfResp)
	require.NoError(t, err)

	require.Equal(t, request.Title, shelfResp.Title)
}

func Test_API_Shelf_Update(t *testing.T) {
	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	shelfId, err := TestService.ShelfService.Create(&model.Shelf{
		ShelfBase: model.ShelfBase{
			Title:       "shelf-title-update",
			Path:        "/shelf-title-update",
			Domain:      "",
			Description: "A shelf created during API integration tests",
			Theme:       "",
			Icon:        "",
			UserId:      userId,
		},
	})
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "shelf-title-updated",
		Path:        "/shelf-title-updated",
		Domain:      "",
		Description: "A shelf updated during API integration tests",
		Theme:       "",
		Icon:        "",
		UserId:      userId,
	}

	resp := doRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/v1/shelves/%s", shelfId),
		strings.NewReader(ObjectToJSON(request)),
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

	var shelfResp model.Shelf
	err = json.Unmarshal(body, &shelfResp)
	require.NoError(t, err)

	require.Equal(t, "shelf-title-updated", shelfResp.Title)
	require.Equal(t, "/shelf-title-updated", shelfResp.Path)
	require.Equal(t, "A shelf updated during API integration tests", shelfResp.Description)
	require.Equal(t, "", shelfResp.Theme)
	require.Equal(t, "", shelfResp.Icon)
	require.Equal(t, shelfId, shelfResp.Id)

}

func Test_API_Shelf_Delete(t *testing.T) {
	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	shelfId, err := TestService.ShelfService.Create(&model.Shelf{
		ShelfBase: model.ShelfBase{
			Title:       "shelf-title-delete",
			Path:        "/shelf-title-delete",
			Domain:      "",
			Description: "A shelf created during API integration tests",
			Theme:       "",
			Icon:        "",
			UserId:      userId,
		},
	})
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/v1/shelves/%s", shelfId),
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

func Test_API_Shelf_Get(t *testing.T) {
	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	shelfId, err := TestService.ShelfService.Create(&model.Shelf{
		ShelfBase: model.ShelfBase{
			Title:       "shelf-title-delete",
			Path:        "/shelf-title-delete",
			Domain:      "",
			Description: "A shelf created during API integration tests",
			Theme:       "",
			Icon:        "",
			UserId:      userId,
		},
	})
	require.NoError(t, err)

	resp := doRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/v1/shelves/%s", shelfId),
		nil,
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

	var shelfResp model.Shelf
	err = json.Unmarshal(body, &shelfResp)
	require.NoError(t, err)

	require.Equal(t, "shelf-title-delete", shelfResp.Title)
	require.Equal(t, "/shelf-title-delete", shelfResp.Path)
	require.Equal(t, "A shelf created during API integration tests", shelfResp.Description)
	require.Equal(t, "", shelfResp.Theme)
	require.Equal(t, "", shelfResp.Icon)
	require.Equal(t, userId, shelfResp.UserId)
}
