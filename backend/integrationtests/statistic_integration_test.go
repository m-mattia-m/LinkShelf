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

func Test_API_Statistic_Get(t *testing.T) {
	_, token := createTestUser(t)

	resp := doAuthedRequest(
		t,
		http.MethodGet,
		"/v1/statistics",
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

	var statistic model.Statistic
	err = json.Unmarshal(body, &statistic)
	require.NoError(t, err)

	require.GreaterOrEqual(t, statistic.ShelfNumber, 0)
	require.GreaterOrEqual(t, statistic.SectionNumber, 0)
	require.GreaterOrEqual(t, statistic.LinkNumber, 0)
}
