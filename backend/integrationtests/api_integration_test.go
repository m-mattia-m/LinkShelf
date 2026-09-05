//go:build integration
// +build integration

package integrationtests

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_API_Readiness(t *testing.T) {
	resp := doRequest(
		t,
		http.MethodGet,
		"/health/readiness",
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

	var responseBody struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(body, &responseBody)
	require.NoError(t, err)

	require.Equal(t, responseBody.Status, "ready")
}

func Test_API_Liveness(t *testing.T) {
	resp := doRequest(
		t,
		http.MethodGet,
		"/health/liveness",
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

	var responseBody struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(body, &responseBody)
	require.NoError(t, err)

	require.Equal(t, responseBody.Status, "alive")
}

func Test_API_Root(t *testing.T) {
	resp := doRequest(
		t,
		http.MethodGet,
		"/",
		nil,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)
	require.Equal(t, http.StatusPermanentRedirect, resp.StatusCode)

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(responseBody), "<a href=\"/swagger\">Permanent Redirect</a>.")

}

func Test_API_SwaggerUI(t *testing.T) {
	resp := doRequest(
		t,
		http.MethodGet,
		"/swagger",
		nil,
	)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	contentType := resp.Header.Get("Content-Type")
	require.Contains(t, contentType, "text/html")
}
