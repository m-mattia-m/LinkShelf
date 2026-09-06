//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_API_Statistic_Get exercises the full HTTP -> domain -> repository ->
// Postgres path for the statistics endpoint.
//
// GetStatistic doesn't yet read the caller's real user id (see the "TODO get
// userId from context/header" in the controller), so it always aggregates for
// the literal placeholder "TODO", which never matches a real, dynamically
// created test user. Because of that this test can only assert that the
// route is wired up, the request succeeds end-to-end against a real
// database, and the response is well-formed - not specific counts driven by
// fixture data. Once the controller reads a real user id this should be
// extended the way Test_API_Shelf_GetPublicByPath_Success is: create known
// shelves/sections/links for a user and assert the exact numbers.
func Test_API_Statistic_Get(t *testing.T) {
	resp := doRequest(
		t,
		http.MethodGet,
		"/v1/statistics",
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

	var statistic model.Statistic
	err = json.Unmarshal(body, &statistic)
	require.NoError(t, err)

	require.GreaterOrEqual(t, statistic.ShelfNumber, 0)
	require.GreaterOrEqual(t, statistic.SectionNumber, 0)
	require.GreaterOrEqual(t, statistic.LinkNumber, 0)
}
