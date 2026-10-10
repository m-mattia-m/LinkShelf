//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const integrationServiceToken = "integration-service-token-0123456789-abcdef"

func enableServiceToken(t *testing.T) {
	t.Helper()
	config.Set("authentication.serviceToken.enabled", true)
	config.Set("authentication.serviceToken.token", integrationServiceToken)
	t.Cleanup(func() {
		config.Set("authentication.serviceToken.enabled", false)
		config.Set("authentication.serviceToken.token", "")
	})
}

func patchLimits(t *testing.T, userId, token, body string) *http.Response {
	t.Helper()
	return doAuthedRequest(t, http.MethodPatch, "/v1/users/"+userId+"/limits", strings.NewReader(body), token)
}

func readLimits(t *testing.T, resp *http.Response) model.UserLimits {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var limits model.UserLimits
	require.NoError(t, json.Unmarshal(raw, &limits))
	return limits
}

func createShelfOverHttp(t *testing.T, token, path string) *http.Response {
	t.Helper()
	return doAuthedRequest(t, http.MethodPost, "/v1/shelves",
		strings.NewReader(ObjectToJSON(model.ShelfBase{Title: path, Path: path})), token)
}

func Test_API_Limits_NewUsersAreUnlimitedByDefault(t *testing.T) {
	userId, token := createTestUser(t)

	resp := doAuthedRequest(t, http.MethodGet, "/v1/users/me", nil, token)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var me model.User
	raw, _ := io.ReadAll(resp.Body)
	require.NoError(t, json.Unmarshal(raw, &me))
	require.Equal(t, userId, me.Id)
	require.Nil(t, me.MaxShelves)
}

func Test_API_Limits_AdminSetsLimitAndShelfCreationIsBlocked(t *testing.T) {
	userId, userToken := createTestUser(t)
	_, adminToken := createTestAdmin(t)

	limits := readLimits(t, patchLimits(t, userId, adminToken, `{"max_shelves": 1}`))
	require.Equal(t, 1, *limits.MaxShelves)

	first := createShelfOverHttp(t, userToken, "limit-"+uniqueUsername())
	_ = first.Body.Close()
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := createShelfOverHttp(t, userToken, "limit-"+uniqueUsername())
	_ = second.Body.Close()
	require.Equal(t, http.StatusForbidden, second.StatusCode)

	// Raising the limit allows creation again, null removes it.
	readLimits(t, patchLimits(t, userId, adminToken, `{"max_shelves": 2}`))
	third := createShelfOverHttp(t, userToken, "limit-"+uniqueUsername())
	_ = third.Body.Close()
	require.Equal(t, http.StatusCreated, third.StatusCode)

	limits = readLimits(t, patchLimits(t, userId, adminToken, `{"max_shelves": null}`))
	require.Nil(t, limits.MaxShelves)
	require.Equal(t, 2, limits.ShelfCount)
}

func Test_API_Limits_LoweringTheLimitKeepsExistingShelves(t *testing.T) {
	userId, userToken := createTestUser(t)
	_, adminToken := createTestAdmin(t)

	for i := 0; i < 2; i++ {
		resp := createShelfOverHttp(t, userToken, fmt.Sprintf("keep-%d-%s", i, uniqueUsername()))
		_ = resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	limits := readLimits(t, patchLimits(t, userId, adminToken, `{"max_shelves": 1}`))
	require.Equal(t, 2, limits.ShelfCount)

	list := doAuthedRequest(t, http.MethodGet, "/v1/shelves", nil, userToken)
	defer func() { _ = list.Body.Close() }()
	var shelves []model.Shelf
	raw, _ := io.ReadAll(list.Body)
	require.NoError(t, json.Unmarshal(raw, &shelves))
	require.Len(t, shelves, 2)
}

func Test_API_Limits_ZeroBlocksAndNegativeIsRejected(t *testing.T) {
	userId, userToken := createTestUser(t)
	_, adminToken := createTestAdmin(t)

	readLimits(t, patchLimits(t, userId, adminToken, `{"max_shelves": 0}`))
	blocked := createShelfOverHttp(t, userToken, "zero-"+uniqueUsername())
	_ = blocked.Body.Close()
	require.Equal(t, http.StatusForbidden, blocked.StatusCode)

	bad := patchLimits(t, userId, adminToken, `{"max_shelves": -1}`)
	defer func() { _ = bad.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, bad.StatusCode)
}

func Test_API_Limits_UsersCannotChangeTheirOwnLimit(t *testing.T) {
	userId, userToken := createTestUser(t)

	resp := patchLimits(t, userId, userToken, `{"max_shelves": 100}`)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	get := doAuthedRequest(t, http.MethodGet, "/v1/users/"+userId+"/limits", nil, userToken)
	_ = get.Body.Close()
	require.Equal(t, http.StatusForbidden, get.StatusCode)
}

func Test_API_Limits_ProfileUpdateCannotChangeTheLimit(t *testing.T) {
	userId, userToken := createTestUser(t)
	_, adminToken := createTestAdmin(t)
	readLimits(t, patchLimits(t, userId, adminToken, `{"max_shelves": 1}`))

	me := doAuthedRequest(t, http.MethodGet, "/v1/users/me", nil, userToken)
	var current model.User
	raw, _ := io.ReadAll(me.Body)
	_ = me.Body.Close()
	require.NoError(t, json.Unmarshal(raw, &current))

	body := map[string]any{
		"email": current.Email, "username": current.Username,
		"first_name": current.FirstName, "last_name": current.LastName,
		"max_shelves": 100,
	}
	put := doAuthedRequest(t, http.MethodPut, "/v1/users/"+userId, strings.NewReader(ObjectToJSON(body)), userToken)
	_ = put.Body.Close()

	limits, err := TestService.UserService.GetLimits(userId)
	require.NoError(t, err)
	require.Equal(t, 1, *limits.MaxShelves)
}

func Test_API_Limits_ServiceToken(t *testing.T) {
	enableServiceToken(t)
	userId, _ := createTestUser(t)

	limits := readLimits(t, patchLimits(t, userId, integrationServiceToken, `{"max_shelves": 3}`))
	require.Equal(t, 3, *limits.MaxShelves)

	got := readLimits(t, doAuthedRequest(t, http.MethodGet, "/v1/users/"+userId+"/limits", nil, integrationServiceToken))
	require.Equal(t, 3, *got.MaxShelves)

	missing := patchLimits(t, "00000000-0000-0000-0000-000000000000", integrationServiceToken, `{"max_shelves": 3}`)
	_ = missing.Body.Close()
	require.Equal(t, http.StatusNotFound, missing.StatusCode)
}

func Test_API_Limits_ServiceTokenIsRejectedOnEveryOtherEndpoint(t *testing.T) {
	enableServiceToken(t)
	userId, _ := createTestUser(t)

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/v1/users"},
		{http.MethodGet, "/v1/users/" + userId},
		{http.MethodGet, "/v1/users/me"},
		{http.MethodDelete, "/v1/users/" + userId},
		{http.MethodPatch, "/v1/users/" + userId + "/verify"},
		{http.MethodGet, "/v1/shelves"},
	} {
		resp := doAuthedRequest(t, tc.method, tc.path, nil, integrationServiceToken)
		_ = resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode, tc.method+" "+tc.path)
	}
}

func Test_API_Limits_ServiceTokenIsRejectedWhenDisabled(t *testing.T) {
	userId, _ := createTestUser(t)
	config.Set("authentication.serviceToken.token", integrationServiceToken)
	t.Cleanup(func() { config.Set("authentication.serviceToken.token", "") })

	resp := patchLimits(t, userId, integrationServiceToken, `{"max_shelves": 3}`)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_API_Limits_UpgradeUrlIsPublishedAndIncludedInTheError(t *testing.T) {
	config.Set("limits.upgradeUrl", "https://account.example.com")
	t.Cleanup(func() { config.Set("limits.upgradeUrl", "") })

	userId, userToken := createTestUser(t)
	_, adminToken := createTestAdmin(t)
	readLimits(t, patchLimits(t, userId, adminToken, `{"max_shelves": 0}`))

	resp := createShelfOverHttp(t, userToken, "upgrade-"+uniqueUsername())
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	raw, _ := io.ReadAll(resp.Body)
	require.Contains(t, string(raw), "https://account.example.com")

	settings := doRequest(t, http.MethodGet, "/v1/settings?language_code=en", nil)
	defer func() { _ = settings.Body.Close() }()
	raw, _ = io.ReadAll(settings.Body)
	require.Contains(t, string(raw), `"upgrade_url":"https://account.example.com"`)
}
