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

func Test_API_Shelf_Create(t *testing.T) {

	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "shelf-title-creation",
		Path:        "shelf-title-creation",
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
		PublicShelf: model.PublicShelf{
			Title:       "shelf-title-update",
			Path:        "shelf-title-update",
			Description: "A shelf created during API integration tests",
			Icon:        "",
		},
		Domain: "",
		Theme:  "",
		UserId: userId,
	})
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "shelf-title-updated",
		Path:        "shelf-title-updated",
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
	require.Equal(t, "shelf-title-updated", shelfResp.Path)
	require.Equal(t, "A shelf updated during API integration tests", shelfResp.Description)
	require.Equal(t, "", shelfResp.Theme)
	require.Equal(t, "", shelfResp.Icon)
	require.Equal(t, shelfId, shelfResp.Id)

}

func Test_API_Shelf_Delete(t *testing.T) {
	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	shelfId, err := TestService.ShelfService.Create(&model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       "shelf-title-delete",
			Path:        "shelf-title-delete",
			Description: "A shelf created during API integration tests",
			Icon:        "",
		},
		Domain: "",
		Theme:  "",
		UserId: userId,
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
		PublicShelf: model.PublicShelf{
			Title:       "shelf-title-get",
			Path:        "shelf-title-get",
			Description: "A shelf created during API integration tests",
			Icon:        "",
		},
		Domain: "",
		Theme:  "",
		UserId: userId,
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

	require.Equal(t, "shelf-title-get", shelfResp.Title)
	require.Equal(t, "shelf-title-get", shelfResp.Path)
	require.Equal(t, "A shelf created during API integration tests", shelfResp.Description)
	require.Equal(t, "", shelfResp.Theme)
	require.Equal(t, "", shelfResp.Icon)
	require.Equal(t, userId, shelfResp.UserId)
}

func Test_API_Shelf_Create_DuplicatePath_Conflict(t *testing.T) {
	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "shelf-title-duplicate-path",
		Path:        "shelf-duplicate-path",
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

	respConflict := doRequest(
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
	}(respConflict.Body)

	require.Equal(t, http.StatusConflict, respConflict.StatusCode)
}

func Test_API_Shelf_Create_MissingTitle_Validation(t *testing.T) {
	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "",
		Path:        "shelf-missing-title",
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

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func Test_API_Shelf_Create_InvalidPath_Validation(t *testing.T) {
	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "shelf-invalid-path",
		Path:        "not a valid path!",
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

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func Test_API_Shelf_GetPublicByPath_Success(t *testing.T) {
	userId, err := getShelfOwnerUser()
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "shelf-public-path",
		Path:        "Shelf-Public-Path",
		Domain:      "",
		Description: "A public shelf description",
		Theme:       "",
		Icon:        "i-lucide-book-open",
		UserId:      userId,
	}

	createResp := doRequest(
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
	}(createResp.Body)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	// path lookup is case-insensitive
	resp := doRequest(
		t,
		http.MethodGet,
		"/v1/shelves/by-path/shelf-public-path",
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

	var publicShelf model.PublicShelf
	err = json.Unmarshal(body, &publicShelf)
	require.NoError(t, err)

	require.Equal(t, "shelf-public-path", publicShelf.Title)
	require.Equal(t, "A public shelf description", publicShelf.Description)
	require.Equal(t, "i-lucide-book-open", publicShelf.Icon)
	require.Equal(t, "Shelf-Public-Path", publicShelf.Path)

	// the public payload must not leak internal fields
	require.NotContains(t, string(body), "userId")
	require.NotContains(t, string(body), "domain")
	require.NotContains(t, string(body), "theme")
}

func Test_API_Shelf_GetPublicByPath_NotFound(t *testing.T) {
	resp := doRequest(
		t,
		http.MethodGet,
		"/v1/shelves/by-path/does-not-exist",
		nil,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
