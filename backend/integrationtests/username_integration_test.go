//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// userBasedPaths switches app.userBasedPaths for one test and restores the
// default afterwards, since the server reads the setting on every request.
func userBasedPaths(t *testing.T, enabled bool) {
	t.Helper()
	config.Set("app.userBasedPaths", enabled)
	t.Cleanup(func() { config.Set("app.userBasedPaths", false) })
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return body
}

func currentUser(t *testing.T, token string) model.User {
	t.Helper()
	resp := doAuthedRequest(t, http.MethodGet, "/v1/users/me", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var user model.User
	require.NoError(t, json.Unmarshal(readBody(t, resp), &user))
	require.NotEmpty(t, user.Username)
	return user
}

// createShelfWithPath creates a shelf through the API and returns its id.
func createShelfWithPath(t *testing.T, token, path string) (id string, status int) {
	t.Helper()
	resp := doAuthedRequest(t, http.MethodPost, "/v1/shelves", strings.NewReader(ObjectToJSON(model.ShelfBase{
		Title: "shelf " + path,
		Path:  path,
	})), token)
	status = resp.StatusCode
	body := readBody(t, resp)
	if status != http.StatusCreated {
		return "", status
	}

	var shelf model.Shelf
	require.NoError(t, json.Unmarshal(body, &shelf))
	return shelf.Id, status
}

func lookup(t *testing.T, url string) (status int, shelf model.PublicShelf) {
	t.Helper()
	resp := doRequest(t, http.MethodGet, url, nil)
	status = resp.StatusCode
	body := readBody(t, resp)
	if status == http.StatusOK {
		require.NoError(t, json.Unmarshal(body, &shelf))
	}
	return status, shelf
}

func uniquePath() string {
	return "p-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
}

func Test_API_Username_CreateUser(t *testing.T) {
	username := uniqueUsername()
	request := model.UserCreate{
		UserBase: model.UserBase{Username: username, Email: username + "@test.com", FirstName: "First", LastName: "Last"},
		Password: "secret",
	}

	resp := doRequest(t, http.MethodPost, "/v1/users", strings.NewReader(ObjectToJSON(request)))
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created model.User
	require.NoError(t, json.Unmarshal(readBody(t, resp), &created))
	require.Equal(t, username, created.Username)
}

func Test_API_Username_CreateUser_Rejections(t *testing.T) {
	taken := uniqueUsername()
	first := model.UserCreate{
		UserBase: model.UserBase{Username: taken, Email: taken + "@test.com", FirstName: "First", LastName: "Last"},
		Password: "secret",
	}
	resp := doRequest(t, http.MethodPost, "/v1/users", strings.NewReader(ObjectToJSON(first)))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = readBody(t, resp)

	cases := map[string]struct {
		username string
		status   int
	}{
		"already taken":        {taken, http.StatusConflict},
		"reserved (routing)":   {"docs", http.StatusBadRequest},
		"reserved (generic)":   {"support", http.StatusBadRequest},
		"reserved (bootstrap)": {"admin", http.StatusBadRequest},
		"uppercase":            {"Mixed-Case", http.StatusUnprocessableEntity},
		"too short":            {"ab", http.StatusUnprocessableEntity},
		"underscore":           {"under_score", http.StatusUnprocessableEntity},
		"leading hyphen":       {"-leading", http.StatusUnprocessableEntity},
		"trailing hyphen":      {"trailing-", http.StatusUnprocessableEntity},
		"missing":              {"", http.StatusUnprocessableEntity},
		"too long":             {strings.Repeat("a", 31), http.StatusUnprocessableEntity},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			request := model.UserCreate{
				UserBase: model.UserBase{Username: c.username, Email: "reject-" + uniqueUsername() + "@test.com", FirstName: "First", LastName: "Last"},
				Password: "secret",
			}

			resp := doRequest(t, http.MethodPost, "/v1/users", strings.NewReader(ObjectToJSON(request)))

			require.Equal(t, c.status, resp.StatusCode, string(readBody(t, resp)))
		})
	}
}

func Test_API_Username_Rename(t *testing.T) {
	userId, token := createTestUser(t)
	user := currentUser(t, token)

	other, otherToken := createTestUser(t)
	_ = other
	otherName := currentUser(t, otherToken).Username

	update := func(username string) *http.Response {
		return doAuthedRequest(t, http.MethodPut, "/v1/users/"+userId, strings.NewReader(ObjectToJSON(model.UserBase{
			Username: username, Email: user.Email, FirstName: user.FirstName, LastName: user.LastName,
		})), token)
	}

	newName := uniqueUsername()
	resp := update(newName)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var renamed model.User
	require.NoError(t, json.Unmarshal(readBody(t, resp), &renamed))
	require.Equal(t, newName, renamed.Username)
	require.Equal(t, newName, currentUser(t, token).Username)

	require.Equal(t, http.StatusConflict, update(otherName).StatusCode)
	require.Equal(t, http.StatusBadRequest, update("docs").StatusCode)
	// Saving with the unchanged name still works.
	require.Equal(t, http.StatusOK, update(newName).StatusCode)
}

func Test_API_Username_SettingsExposeTheFlag(t *testing.T) {
	read := func() bool {
		resp := doRequest(t, http.MethodGet, "/v1/settings?language_code=en", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var settings model.SettingPageBody
		require.NoError(t, json.Unmarshal(readBody(t, resp), &settings))
		return settings.UserBasedPaths
	}

	require.False(t, read(), "off by default")

	userBasedPaths(t, true)
	require.True(t, read())
}

func Test_API_UserBasedPaths_Off_PathsAreUniqueAcrossTheInstance(t *testing.T) {
	userBasedPaths(t, false)
	_, aliceToken := createTestUser(t)
	_, bobToken := createTestUser(t)
	path := uniquePath()

	aliceShelf, status := createShelfWithPath(t, aliceToken, path)
	require.Equal(t, http.StatusCreated, status)

	_, status = createShelfWithPath(t, bobToken, path)
	require.Equal(t, http.StatusConflict, status, "another user cannot take /"+path)

	// The lookup is case-insensitive, so a different casing collides too.
	_, status = createShelfWithPath(t, bobToken, strings.ToUpper(path))
	require.Equal(t, http.StatusConflict, status)

	status, shelf := lookup(t, "/v1/shelves/by-path/"+path)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, aliceShelf, shelf.Id)

	// The username lookup does not answer while the feature is off.
	status, _ = lookup(t, "/v1/shelves/by-user/"+currentUser(t, aliceToken).Username+"/"+path)
	require.Equal(t, http.StatusNotFound, status)
}

func Test_API_UserBasedPaths_Off_RouteWordsCannotBeShelfPaths(t *testing.T) {
	userBasedPaths(t, false)
	_, token := createTestUser(t)

	for _, path := range []string{"app", "auth", "docs", "cloud", "about", "contact", "imprint", "privacy-policy", "terms-of-use", "api", "v1", "swagger", "health", "images"} {
		_, status := createShelfWithPath(t, token, path)
		require.Equal(t, http.StatusBadRequest, status, path)
	}

	// Words that are only reserved as usernames are fine as a path.
	path := "support-" + uniquePath()
	_, status := createShelfWithPath(t, token, path)
	require.Equal(t, http.StatusCreated, status)
}

func Test_API_UserBasedPaths_On_TwoUsersCanShareAPath(t *testing.T) {
	userBasedPaths(t, true)
	_, aliceToken := createTestUser(t)
	_, bobToken := createTestUser(t)
	alice := currentUser(t, aliceToken).Username
	bob := currentUser(t, bobToken).Username
	path := uniquePath()

	aliceShelf, status := createShelfWithPath(t, aliceToken, path)
	require.Equal(t, http.StatusCreated, status)
	bobShelf, status := createShelfWithPath(t, bobToken, path)
	require.Equal(t, http.StatusCreated, status, "the same path is fine behind a different username")
	require.NotEqual(t, aliceShelf, bobShelf)

	// A user cannot have the same path twice, though.
	_, status = createShelfWithPath(t, aliceToken, path)
	require.Equal(t, http.StatusConflict, status)

	status, shelf := lookup(t, "/v1/shelves/by-user/"+alice+"/"+path)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, aliceShelf, shelf.Id)

	status, shelf = lookup(t, "/v1/shelves/by-user/"+bob+"/"+path)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, bobShelf, shelf.Id)

	// A username is not case sensitive in the URL.
	status, shelf = lookup(t, "/v1/shelves/by-user/"+strings.ToUpper(alice)+"/"+path)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, aliceShelf, shelf.Id)

	// Without the username the path no longer resolves.
	status, _ = lookup(t, "/v1/shelves/by-path/"+path)
	require.Equal(t, http.StatusNotFound, status)

	// Nobody else's username finds the shelf either.
	status, _ = lookup(t, "/v1/shelves/by-user/"+bob+"/does-not-exist")
	require.Equal(t, http.StatusNotFound, status)
}

func Test_API_UserBasedPaths_On_RouteWordsAreFineAsAPath(t *testing.T) {
	userBasedPaths(t, true)
	_, token := createTestUser(t)
	username := currentUser(t, token).Username

	id, status := createShelfWithPath(t, token, "docs")
	require.Equal(t, http.StatusCreated, status)

	status, shelf := lookup(t, "/v1/shelves/by-user/"+username+"/docs")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, id, shelf.Id)
}

func Test_API_UserBasedPaths_On_RenamingMovesEveryShelfAndBreaksOldLinks(t *testing.T) {
	userBasedPaths(t, true)
	userId, token := createTestUser(t)
	user := currentUser(t, token)
	oldName := user.Username
	first, _ := createShelfWithPath(t, token, "first-"+uniquePath())
	second, _ := createShelfWithPath(t, token, "second-"+uniquePath())

	var firstPath, secondPath string
	for id, dst := range map[string]*string{first: &firstPath, second: &secondPath} {
		resp := doAuthedRequest(t, http.MethodGet, "/v1/shelves/"+id, nil, token)
		var shelf model.Shelf
		require.NoError(t, json.Unmarshal(readBody(t, resp), &shelf))
		require.Equal(t, oldName, shelf.Username, "the owner's username is part of the shelf response")
		*dst = shelf.Path
	}

	newName := uniqueUsername()
	resp := doAuthedRequest(t, http.MethodPut, "/v1/users/"+userId, strings.NewReader(ObjectToJSON(model.UserBase{
		Username: newName, Email: user.Email, FirstName: user.FirstName, LastName: user.LastName,
	})), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = readBody(t, resp)

	for id, path := range map[string]string{first: firstPath, second: secondPath} {
		// The old link is gone ...
		status, _ := lookup(t, "/v1/shelves/by-user/"+oldName+"/"+path)
		require.Equal(t, http.StatusNotFound, status)

		// ... and the shelf is found under the new name.
		status, shelf := lookup(t, "/v1/shelves/by-user/"+newName+"/"+path)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, id, shelf.Id)
	}

	// The old name is free again.
	other := model.UserCreate{
		UserBase: model.UserBase{Username: oldName, Email: "reuse-" + uniqueUsername() + "@test.com", FirstName: "F", LastName: "L"},
		Password: "secret",
	}
	resp = doRequest(t, http.MethodPost, "/v1/users", strings.NewReader(ObjectToJSON(other)))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = readBody(t, resp)
}

func Test_API_UserBasedPaths_Off_AnExistingReservedPathCanStillBeEdited(t *testing.T) {
	userBasedPaths(t, true)
	userId, token := createTestUser(t)
	id, status := createShelfWithPath(t, token, "docs")
	require.Equal(t, http.StatusCreated, status)

	// Switching the feature off leaves this shelf on a path that is now
	// reserved. Its title can still be changed, but the path cannot be moved
	// onto another reserved word.
	userBasedPaths(t, false)
	_ = userId

	update := func(title, path string) int {
		resp := doAuthedRequest(t, http.MethodPut, "/v1/shelves/"+id, strings.NewReader(ObjectToJSON(model.ShelfBase{Title: title, Path: path})), token)
		_ = readBody(t, resp)
		return resp.StatusCode
	}

	require.Equal(t, http.StatusOK, update("renamed title", "docs"))
	require.Equal(t, http.StatusBadRequest, update("renamed title", "app"))
	require.Equal(t, http.StatusOK, update("renamed title", uniquePath()))
}

func getShelf(t *testing.T, token, id string) model.Shelf {
	t.Helper()
	resp := doAuthedRequest(t, http.MethodGet, "/v1/shelves/"+id, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var shelf model.Shelf
	require.NoError(t, json.Unmarshal(readBody(t, resp), &shelf))
	return shelf
}

func Test_API_Shelf_RemembersWhichModeItWasCreatedIn(t *testing.T) {
	_, token := createTestUser(t)

	userBasedPaths(t, false)
	createdOff, status := createShelfWithPath(t, token, uniquePath())
	require.Equal(t, http.StatusCreated, status)

	userBasedPaths(t, true)
	createdOn, status := createShelfWithPath(t, token, uniquePath())
	require.Equal(t, http.StatusCreated, status)

	// Read back under each setting: what a shelf was created under never moves.
	for _, enabled := range []bool{true, false, true} {
		userBasedPaths(t, enabled)
		require.False(t, getShelf(t, token, createdOff).CreatedWithUserBasedPaths, "created while off (setting now %v)", enabled)
		require.True(t, getShelf(t, token, createdOn).CreatedWithUserBasedPaths, "created while on (setting now %v)", enabled)
	}
}

func Test_API_Shelf_EditingDoesNotChangeTheCreationMode(t *testing.T) {
	_, token := createTestUser(t)
	userBasedPaths(t, false)
	id, status := createShelfWithPath(t, token, uniquePath())
	require.Equal(t, http.StatusCreated, status)

	userBasedPaths(t, true)
	resp := doAuthedRequest(t, http.MethodPut, "/v1/shelves/"+id, strings.NewReader(ObjectToJSON(model.ShelfBase{
		Title: "edited under user-based paths", Path: uniquePath(),
	})), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var updated model.Shelf
	require.NoError(t, json.Unmarshal(readBody(t, resp), &updated))

	require.False(t, updated.CreatedWithUserBasedPaths, "a new title and path do not make it a user-based shelf")
	require.False(t, getShelf(t, token, id).CreatedWithUserBasedPaths)
}

func Test_API_Shelf_CreationModeCannotBeSetByTheClient(t *testing.T) {
	_, token := createTestUser(t)
	userBasedPaths(t, false)

	body := `{"title":"sneaky","path":"` + uniquePath() + `","createdWithUserBasedPaths":true}`
	resp := doAuthedRequest(t, http.MethodPost, "/v1/shelves", strings.NewReader(body), token)
	// Either the extra field is refused or it is ignored - it must never stick.
	if resp.StatusCode == http.StatusCreated {
		var created model.Shelf
		require.NoError(t, json.Unmarshal(readBody(t, resp), &created))
		require.False(t, created.CreatedWithUserBasedPaths)
		return
	}
	_ = readBody(t, resp)
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}
