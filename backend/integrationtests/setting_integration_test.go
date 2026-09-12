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

func Test_API_Setting_UpdateSettingsBatch_RequiresAdmin(t *testing.T) {
	_, userToken := createTestUser(t)

	resp := doAuthedRequest(t, http.MethodPut, "/v1/settings/batch", strings.NewReader(ObjectToJSON(model.SettingBatchRequestBody{
		Settings: []model.Setting{{Key: "about_show", LanguageCode: "en", Value: "true"}},
	})), userToken)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func Test_API_Setting_UpdateSettingsBatch_Success(t *testing.T) {
	_, adminToken := createTestAdmin(t)

	resp := doAuthedRequest(t, http.MethodPut, "/v1/settings/batch", strings.NewReader(ObjectToJSON(model.SettingBatchRequestBody{
		Settings: []model.Setting{
			{Key: "contact", LanguageCode: "en", Value: "hello@example.com"},
			{Key: "contact_show", LanguageCode: "en", Value: "true"},
		},
	})), adminToken)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp model.SettingBatchResponseBody
	require.NoError(t, json.Unmarshal(body, &batchResp))
	require.Empty(t, batchResp.Failures)
	require.Equal(t, "hello@example.com", batchResp.Settings.Contact)
	require.True(t, batchResp.Settings.ContactShow)

	// Confirm it actually persisted, not just reflected in the response.
	getResp := doRequest(t, http.MethodGet, "/v1/settings?language_code=en", nil)
	defer func() { _ = getResp.Body.Close() }()
	getBody, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)
	var reloaded model.SettingPageBody
	require.NoError(t, json.Unmarshal(getBody, &reloaded))
	require.Equal(t, "hello@example.com", reloaded.Contact)
	require.True(t, reloaded.ContactShow)
}

func Test_API_Setting_UpdateSettingsBatch_PartialFailureStillSavesValidItems(t *testing.T) {
	_, adminToken := createTestAdmin(t)

	resp := doAuthedRequest(t, http.MethodPut, "/v1/settings/batch", strings.NewReader(ObjectToJSON(model.SettingBatchRequestBody{
		Settings: []model.Setting{
			{Key: "imprint", LanguageCode: "en", Value: "Imprint text"},
			{Key: "imprint_show", LanguageCode: "en", Value: "not-a-bool"},
			{Key: "not_a_real_key", LanguageCode: "en", Value: "x"},
		},
	})), adminToken)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp model.SettingBatchResponseBody
	require.NoError(t, json.Unmarshal(body, &batchResp))
	require.Len(t, batchResp.Failures, 2)
	require.Equal(t, "Imprint text", batchResp.Settings.Imprint)

	// The valid item must have actually persisted despite the other two failing.
	getResp := doRequest(t, http.MethodGet, "/v1/settings?language_code=en", nil)
	defer func() { _ = getResp.Body.Close() }()
	getBody, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)
	var reloaded model.SettingPageBody
	require.NoError(t, json.Unmarshal(getBody, &reloaded))
	require.Equal(t, "Imprint text", reloaded.Imprint)
}

func Test_API_Setting_GetEmailDeliveryInfo_RequiresAdmin(t *testing.T) {
	_, userToken := createTestUser(t)

	resp := doAuthedRequest(t, http.MethodGet, "/v1/settings/email-delivery", nil, userToken)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func Test_API_Setting_GetEmailDeliveryInfo_Success(t *testing.T) {
	config.Set("smtp.host", "smtp.example.com")
	config.Set("smtp.from", "no-reply@example.com")
	defer func() {
		config.Set("smtp.host", "")
		config.Set("smtp.from", "")
	}()

	_, adminToken := createTestAdmin(t)

	resp := doAuthedRequest(t, http.MethodGet, "/v1/settings/email-delivery", nil, adminToken)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var info model.EmailDeliveryInfo
	require.NoError(t, json.Unmarshal(body, &info))
	require.Equal(t, "smtp.example.com", info.Host)
	require.Equal(t, "no-reply@example.com", info.From)
}

func Test_API_Setting_UpdateSettingsBatch_EmptyBatchRejected(t *testing.T) {
	_, adminToken := createTestAdmin(t)

	resp := doAuthedRequest(t, http.MethodPut, "/v1/settings/batch", strings.NewReader(ObjectToJSON(model.SettingBatchRequestBody{
		Settings: []model.Setting{},
	})), adminToken)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
