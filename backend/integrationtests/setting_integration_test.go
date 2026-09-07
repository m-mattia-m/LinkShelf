//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_API_Setting_GetPageSettings_Public(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/settings?language_code=en", nil)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var settings model.SettingPageBody
	require.NoError(t, json.Unmarshal(body, &settings))
	// Seeded by migrations/*/0002_data.up.sql - proves settings actually load
	// from the real database, not just that the route responds.
	require.Contains(t, settings.About, "LinkShelf")
}

func Test_API_Setting_UpdateSetting_RequiresAdmin(t *testing.T) {
	_, userToken := createTestUser(t)

	resp := doAuthedRequest(t, http.MethodPut, "/v1/settings", strings.NewReader(ObjectToJSON(model.Setting{
		Key: "about_show", LanguageCode: "en", Value: "true",
	})), userToken)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func Test_API_Setting_UpdateSetting_Success(t *testing.T) {
	_, adminToken := createTestAdmin(t)

	updateResp := doAuthedRequest(t, http.MethodPut, "/v1/settings", strings.NewReader(ObjectToJSON(model.Setting{
		Key: "redirect_to_dashboard", LanguageCode: "en", Value: "true",
	})), adminToken)
	defer func() { _ = updateResp.Body.Close() }()

	require.Equal(t, http.StatusOK, updateResp.StatusCode)

	body, err := io.ReadAll(updateResp.Body)
	require.NoError(t, err)

	var settings model.SettingPageBody
	require.NoError(t, json.Unmarshal(body, &settings))
	require.True(t, settings.RedirectToDashboard)

	// Confirm it actually persisted, not just reflected in the response.
	getResp := doRequest(t, http.MethodGet, "/v1/settings?language_code=en", nil)
	defer func() { _ = getResp.Body.Close() }()
	getBody, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)
	var reloaded model.SettingPageBody
	require.NoError(t, json.Unmarshal(getBody, &reloaded))
	require.True(t, reloaded.RedirectToDashboard)
}
