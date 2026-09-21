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

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_API_Shelf_Create(t *testing.T) {
	_, token := createTestUser(t)

	request := model.ShelfBase{
		Title:       "shelf-title-creation",
		Path:        "shelf-title-creation",
		Domain:      "",
		Description: "A shelf created during API integration tests",
		ThemeId:     "",
		Icon:        "",
	}

	resp := doAuthedRequest(
		t,
		http.MethodPost,
		"/v1/shelves",
		strings.NewReader(ObjectToJSON(request)),
		token,
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
	userId, token := createTestUser(t)

	shelfId, err := TestService.ShelfService.Create(userId, &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       "shelf-title-update",
			Path:        "shelf-title-update",
			Description: "A shelf created during API integration tests",
			Icon:        "",
		},
		Domain:  "",
		ThemeId: "",
	})
	require.NoError(t, err)

	request := model.ShelfBase{
		Title:       "shelf-title-updated",
		Path:        "shelf-title-updated",
		Domain:      "",
		Description: "A shelf updated during API integration tests",
		ThemeId:     "",
		Icon:        "",
	}

	resp := doAuthedRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/v1/shelves/%s", shelfId),
		strings.NewReader(ObjectToJSON(request)),
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

	var shelfResp model.Shelf
	err = json.Unmarshal(body, &shelfResp)
	require.NoError(t, err)

	require.Equal(t, "shelf-title-updated", shelfResp.Title)
	require.Equal(t, "shelf-title-updated", shelfResp.Path)
	require.Equal(t, "A shelf updated during API integration tests", shelfResp.Description)
	require.Equal(t, "", shelfResp.ThemeId)
	require.Equal(t, "", shelfResp.Icon)
	require.Equal(t, shelfId, shelfResp.Id)

}

func Test_API_Shelf_Delete(t *testing.T) {
	userId, token := createTestUser(t)

	shelfId, err := TestService.ShelfService.Create(userId, &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       "shelf-title-delete",
			Path:        "shelf-title-delete",
			Description: "A shelf created during API integration tests",
			Icon:        "",
		},
		Domain:  "",
		ThemeId: "",
	})
	require.NoError(t, err)

	resp := doAuthedRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/v1/shelves/%s", shelfId),
		nil,
		token,
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
	userId, token := createTestUser(t)

	shelfId, err := TestService.ShelfService.Create(userId, &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       "shelf-title-get",
			Path:        "shelf-title-get",
			Description: "A shelf created during API integration tests",
			Icon:        "",
		},
		Domain:  "",
		ThemeId: "",
	})
	require.NoError(t, err)

	resp := doAuthedRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/v1/shelves/%s", shelfId),
		nil,
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

	var shelfResp model.Shelf
	err = json.Unmarshal(body, &shelfResp)
	require.NoError(t, err)

	require.Equal(t, "shelf-title-get", shelfResp.Title)
	require.Equal(t, "shelf-title-get", shelfResp.Path)
	require.Equal(t, "A shelf created during API integration tests", shelfResp.Description)
	require.Equal(t, "", shelfResp.ThemeId)
	require.Equal(t, "", shelfResp.Icon)
	require.Equal(t, userId, shelfResp.UserId)
}

func Test_API_Shelf_Create_DuplicatePath_Conflict(t *testing.T) {
	_, token := createTestUser(t)

	request := model.ShelfBase{
		Title:       "shelf-title-duplicate-path",
		Path:        "shelf-duplicate-path",
		Domain:      "",
		Description: "A shelf created during API integration tests",
		ThemeId:     "",
		Icon:        "",
	}

	resp := doAuthedRequest(
		t,
		http.MethodPost,
		"/v1/shelves",
		strings.NewReader(ObjectToJSON(request)),
		token,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	respConflict := doAuthedRequest(
		t,
		http.MethodPost,
		"/v1/shelves",
		strings.NewReader(ObjectToJSON(request)),
		token,
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
	_, token := createTestUser(t)

	request := model.ShelfBase{
		Title:       "",
		Path:        "shelf-missing-title",
		Domain:      "",
		Description: "A shelf created during API integration tests",
		ThemeId:     "",
		Icon:        "",
	}

	resp := doAuthedRequest(
		t,
		http.MethodPost,
		"/v1/shelves",
		strings.NewReader(ObjectToJSON(request)),
		token,
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
	_, token := createTestUser(t)

	request := model.ShelfBase{
		Title:       "shelf-invalid-path",
		Path:        "not a valid path!",
		Domain:      "",
		Description: "A shelf created during API integration tests",
		ThemeId:     "",
		Icon:        "",
	}

	resp := doAuthedRequest(
		t,
		http.MethodPost,
		"/v1/shelves",
		strings.NewReader(ObjectToJSON(request)),
		token,
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
	_, token := createTestUser(t)

	request := model.ShelfBase{
		Title:       "shelf-public-path",
		Path:        "Shelf-Public-Path",
		Domain:      "",
		Description: "A public shelf description",
		ThemeId:     "",
		Icon:        "i-lucide-book-open",
	}

	createResp := doAuthedRequest(
		t,
		http.MethodPost,
		"/v1/shelves",
		strings.NewReader(ObjectToJSON(request)),
		token,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(createResp.Body)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	// path lookup is case-insensitive, and requires no authentication
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
	// no theme was selected, so it resolves to nil (render the built-in
	// default look) rather than leaking the internal theme_id.
	require.Nil(t, publicShelf.Theme)

	// the public payload must not leak internal fields
	require.NotContains(t, string(body), "userId")
	require.NotContains(t, string(body), "domain")
	require.NotContains(t, string(body), "themeId")
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

// uniqueSuffix is a short lowercase token that is valid inside a domain label
// and a path, so tests that share one database can't collide.
func uniqueSuffix() string {
	return strings.ToLower(uuid.NewString()[:8])
}

// createShelfRequest posts a shelf and returns the response, closing nothing:
// the caller decides what to read.
func createShelfRequest(t *testing.T, token string, shelf model.ShelfBase) *http.Response {
	t.Helper()

	return doAuthedRequest(
		t,
		http.MethodPost,
		"/v1/shelves",
		strings.NewReader(ObjectToJSON(shelf)),
		token,
	)
}

func Test_API_Shelf_Domain_CreateAndPublicLookup(t *testing.T) {
	_, token := createTestUser(t)
	suffix := uniqueSuffix()

	// A domain shelf has no path, and the domain is stored normalized.
	resp := createShelfRequest(t, token, model.ShelfBase{
		Title:       "shelf-domain",
		Domain:      "  Profile-" + suffix + ".Example.com:9443/ ",
		Description: "Served on a host of its own",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var created model.Shelf
	require.NoError(t, json.Unmarshal(body, &created))
	require.Equal(t, "profile-"+suffix+".example.com:9443", created.Domain)
	require.Empty(t, created.Path)

	// The public lookup finds it by any spelling of the domain, and exposes
	// nothing but the public fields.
	lookup := doRequest(t, http.MethodGet, "/v1/shelves/by-domain/Profile-"+suffix+".example.com%3A9443", nil)
	defer lookup.Body.Close()
	require.Equal(t, http.StatusOK, lookup.StatusCode)

	lookupBody, err := io.ReadAll(lookup.Body)
	require.NoError(t, err)
	var publicShelf model.PublicShelf
	require.NoError(t, json.Unmarshal(lookupBody, &publicShelf))
	require.Equal(t, created.Id, publicShelf.Id)
	require.Equal(t, "shelf-domain", publicShelf.Title)
	require.NotContains(t, string(lookupBody), "userId")
	require.NotContains(t, string(lookupBody), `"domain"`)

	// Another port is another site.
	other := doRequest(t, http.MethodGet, "/v1/shelves/by-domain/profile-"+suffix+".example.com", nil)
	defer other.Body.Close()
	require.Equal(t, http.StatusNotFound, other.StatusCode)
}

func Test_API_Shelf_Domain_DefaultPortIsStripped(t *testing.T) {
	_, token := createTestUser(t)
	domain := "port-" + uniqueSuffix() + ".example.com"

	resp := createShelfRequest(t, token, model.ShelfBase{Title: "shelf-default-port", Domain: domain + ":443"})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	lookup := doRequest(t, http.MethodGet, "/v1/shelves/by-domain/"+domain, nil)
	defer lookup.Body.Close()
	require.Equal(t, http.StatusOK, lookup.StatusCode)
}

func Test_API_Shelf_Domain_DuplicateIsConflict_CaseInsensitive(t *testing.T) {
	_, token := createTestUser(t)
	_, otherToken := createTestUser(t)
	domain := "dup-" + uniqueSuffix() + ".example.com"

	first := createShelfRequest(t, token, model.ShelfBase{Title: "shelf-domain-first", Domain: domain})
	defer first.Body.Close()
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := createShelfRequest(t, otherToken, model.ShelfBase{Title: "shelf-domain-second", Domain: strings.ToUpper(domain) + "."})
	defer second.Body.Close()
	require.Equal(t, http.StatusConflict, second.StatusCode)
}

func Test_API_Shelf_Domain_RejectsWhatIsNotADomain(t *testing.T) {
	_, token := createTestUser(t)

	for _, bad := range []string{"localhost", "1.2.3.4", "*.example.com", "https://profile.example.com", "profile.example.com/path", "profile.example.com:99999", "profile_x.example.com"} {
		resp := createShelfRequest(t, token, model.ShelfBase{Title: "shelf-bad-domain", Domain: bad})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, bad)
		resp.Body.Close()
	}
}

func Test_API_Shelf_Domain_ExactlyOneOfPathAndDomain(t *testing.T) {
	_, token := createTestUser(t)

	both := createShelfRequest(t, token, model.ShelfBase{Title: "shelf-both", Path: "shelf-both-" + uniqueSuffix(), Domain: "both-" + uniqueSuffix() + ".example.com"})
	defer both.Body.Close()
	require.Equal(t, http.StatusBadRequest, both.StatusCode)

	neither := createShelfRequest(t, token, model.ShelfBase{Title: "shelf-neither"})
	defer neither.Body.Close()
	require.Equal(t, http.StatusBadRequest, neither.StatusCode)
}

func Test_API_Shelf_Domain_SwitchBetweenPathAndDomain(t *testing.T) {
	userId, token := createTestUser(t)
	suffix := uniqueSuffix()

	shelfId, err := TestService.ShelfService.Create(userId, &model.Shelf{
		PublicShelf: model.PublicShelf{Title: "shelf-switch", Path: "shelf-switch-" + suffix},
	})
	require.NoError(t, err)

	update := func(shelf model.ShelfBase) *http.Response {
		return doAuthedRequest(t, http.MethodPut, fmt.Sprintf("/v1/shelves/%s", shelfId), strings.NewReader(ObjectToJSON(shelf)), token)
	}

	// Path -> domain: the path has to be cleared in the same request.
	toDomain := update(model.ShelfBase{Title: "shelf-switch", Domain: "switch-" + suffix + ".example.com"})
	defer toDomain.Body.Close()
	require.Equal(t, http.StatusOK, toDomain.StatusCode)

	byPath := doRequest(t, http.MethodGet, "/v1/shelves/by-path/shelf-switch-"+suffix, nil)
	defer byPath.Body.Close()
	require.Equal(t, http.StatusNotFound, byPath.StatusCode, "the old path no longer resolves")

	byDomain := doRequest(t, http.MethodGet, "/v1/shelves/by-domain/switch-"+suffix+".example.com", nil)
	defer byDomain.Body.Close()
	require.Equal(t, http.StatusOK, byDomain.StatusCode)

	// Keeping both is refused.
	keepBoth := update(model.ShelfBase{Title: "shelf-switch", Path: "shelf-switch-" + suffix, Domain: "switch-" + suffix + ".example.com"})
	defer keepBoth.Body.Close()
	require.Equal(t, http.StatusBadRequest, keepBoth.StatusCode)

	// Domain -> path again frees the domain.
	toPath := update(model.ShelfBase{Title: "shelf-switch", Path: "shelf-switch-" + suffix})
	defer toPath.Body.Close()
	require.Equal(t, http.StatusOK, toPath.StatusCode)

	freed := doRequest(t, http.MethodGet, "/v1/shelves/by-domain/switch-"+suffix+".example.com", nil)
	defer freed.Body.Close()
	require.Equal(t, http.StatusNotFound, freed.StatusCode)
}

func Test_API_Shelf_Domain_UpdateDuplicateIsConflict(t *testing.T) {
	userId, token := createTestUser(t)
	suffix := uniqueSuffix()

	_, err := TestService.ShelfService.Create(userId, &model.Shelf{
		PublicShelf: model.PublicShelf{Title: "shelf-holder"},
		Domain:      "holder-" + suffix + ".example.com",
	})
	require.NoError(t, err)
	shelfId, err := TestService.ShelfService.Create(userId, &model.Shelf{
		PublicShelf: model.PublicShelf{Title: "shelf-wants", Path: "shelf-wants-" + suffix},
	})
	require.NoError(t, err)

	resp := doAuthedRequest(t, http.MethodPut, fmt.Sprintf("/v1/shelves/%s", shelfId),
		strings.NewReader(ObjectToJSON(model.ShelfBase{Title: "shelf-wants", Domain: "holder-" + suffix + ".example.com"})), token)
	defer resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func Test_API_Shelf_Domain_LookupOfUnknownOrInvalidHost(t *testing.T) {
	for _, host := range []string{"unknown-" + uniqueSuffix() + ".example.com", "localhost", "1.2.3.4"} {
		resp := doRequest(t, http.MethodGet, "/v1/shelves/by-domain/"+host, nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode, host)
		resp.Body.Close()
	}
}
