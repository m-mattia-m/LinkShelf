package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/model"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newServiceDomainTestAPI(t *testing.T) (humatest.TestAPI, *MockService) {
	t.Helper()
	svc := NewMockDomainService(t)
	t.Cleanup(svc.Ctrl.Finish)

	config.Reset()
	config.Set("authentication.jwtSecret", "test-secret")
	config.Set("authentication.serviceToken.enabled", true)
	config.Set("authentication.serviceToken.token", testServiceToken)
	config.Set("authentication.serviceToken.scopes", []string{config.ScopeDomainsRead, config.ScopeDomainsWriteStatus})

	_, api := humatest.New(t)
	api.UseMiddleware(NewAuthenticationMiddleware(api))
	huma.Register(api, huma.Operation{
		Method: http.MethodGet, OperationID: "list-service-domains", Path: "/v1/service/domains",
		Security: bearerSecurity(), Metadata: requireServiceScope(config.ScopeDomainsRead), Errors: []int{http.StatusNotModified},
	}, ListServiceDomains(svc.Service))
	huma.Register(api, huma.Operation{
		Method: http.MethodPut, OperationID: "set-service-domain-status", Path: "/v1/service/domains/{domain}/status",
		Security: bearerSecurity(), Metadata: requireServiceScope(config.ScopeDomainsWriteStatus),
	}, SetServiceDomainStatus(svc.Service))
	return api, svc
}

const serviceAuth = "Authorization: Bearer " + testServiceToken

func Test_API_ListServiceDomains_PagesAndAnswers304(t *testing.T) {
	api, svc := newServiceDomainTestAPI(t)

	items := []model.ServiceDomain{{Domain: "a.example.com", ShelfId: "s-1", OwnerId: "u-1", Status: model.ServiceDomainStatus{Records: []model.DomainRecord{}, ObservedDomain: "a.example.com"}}}
	svc.ShelfService.EXPECT().ListDomains("", 100).Times(3).Return(items, "s-1", nil)
	svc.ShelfService.EXPECT().ListDomains("s-1", 2).Return(nil, "", nil)

	resp := api.Get("/v1/service/domains", serviceAuth)
	require.Equal(t, http.StatusOK, resp.Code)
	etag := resp.Header().Get("ETag")
	require.NotEmpty(t, etag)

	var page model.ServiceDomainList
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &page))
	require.Len(t, page.Items, 1)
	require.Equal(t, "s-1", *page.NextCursor)

	// Unchanged: no body, the poll is nearly free.
	resp = api.Get("/v1/service/domains", serviceAuth, "If-None-Match: "+etag)
	require.Equal(t, http.StatusNotModified, resp.Code)
	require.Empty(t, resp.Body.String())

	// A weak validator or a list of them matches too, a different one does not.
	resp = api.Get("/v1/service/domains", serviceAuth, `If-None-Match: "other", W/`+etag)
	require.Equal(t, http.StatusNotModified, resp.Code)

	// Last page: null cursor.
	resp = api.Get("/v1/service/domains?cursor=s-1&limit=2", serviceAuth)
	require.Equal(t, http.StatusOK, resp.Code)
	require.Contains(t, resp.Body.String(), `"next_cursor":null`)
}

func Test_API_ListServiceDomains_ChangedPageIsSentAgain(t *testing.T) {
	api, svc := newServiceDomainTestAPI(t)

	gomock.InOrder(
		svc.ShelfService.EXPECT().ListDomains("", 100).Return([]model.ServiceDomain{{Domain: "a.example.com"}}, "", nil),
		svc.ShelfService.EXPECT().ListDomains("", 100).Return([]model.ServiceDomain{{Domain: "b.example.com"}}, "", nil),
	)

	etag := api.Get("/v1/service/domains", serviceAuth).Header().Get("ETag")
	resp := api.Get("/v1/service/domains", serviceAuth, "If-None-Match: "+etag)

	require.Equal(t, http.StatusOK, resp.Code)
	require.NotEqual(t, etag, resp.Header().Get("ETag"))
}

func Test_API_ListServiceDomains_RejectsLimitsOutOfRange(t *testing.T) {
	api, _ := newServiceDomainTestAPI(t)

	require.Equal(t, http.StatusUnprocessableEntity, api.Get("/v1/service/domains?limit=0", serviceAuth).Code)
	require.Equal(t, http.StatusUnprocessableEntity, api.Get("/v1/service/domains?limit=501", serviceAuth).Code)
}

func Test_API_SetServiceDomainStatus(t *testing.T) {
	body := map[string]any{
		"shelf_id": "s-1", "state": "pending", "reason": "txt_missing",
		"records": []map[string]string{{"type": "TXT", "name": "_c.a.example.com", "value": "v"}},
	}

	t.Run("stores it", func(t *testing.T) {
		api, svc := newServiceDomainTestAPI(t)
		svc.ShelfService.EXPECT().SetDomainStatus("a.example.com:9443", "s-1", gomock.Any()).DoAndReturn(
			func(_, _ string, st model.DomainStatus) (*model.ServiceDomain, error) {
				require.Equal(t, "txt_missing", st.Reason)
				require.Equal(t, "_c.a.example.com", st.Records[0].Name)
				return &model.ServiceDomain{Domain: "a.example.com:9443", ShelfId: "s-1", Status: model.ServiceDomainStatus{State: st.State}}, nil
			})

		resp := api.Put("/v1/service/domains/a.example.com:9443/status", serviceAuth, body)

		require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
		require.Contains(t, resp.Body.String(), `"state":"pending"`)
	})

	t.Run("a stale controller gets 409", func(t *testing.T) {
		api, svc := newServiceDomainTestAPI(t)
		svc.ShelfService.EXPECT().SetDomainStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, domain.ErrConflict)

		require.Equal(t, http.StatusConflict, api.Put("/v1/service/domains/a.example.com/status", serviceAuth, body).Code)
	})

	t.Run("invalid input is 400", func(t *testing.T) {
		api, svc := newServiceDomainTestAPI(t)
		svc.ShelfService.EXPECT().SetDomainStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, domain.ErrInvalidInput)

		require.Equal(t, http.StatusBadRequest, api.Put("/v1/service/domains/a.example.com/status", serviceAuth, body).Code)
	})

	t.Run("a state outside the enum never reaches the service", func(t *testing.T) {
		api, _ := newServiceDomainTestAPI(t)

		resp := api.Put("/v1/service/domains/a.example.com/status", serviceAuth, map[string]any{"shelf_id": "s-1", "state": "done"})
		require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	})

	t.Run("shelf_id is required", func(t *testing.T) {
		api, _ := newServiceDomainTestAPI(t)

		resp := api.Put("/v1/service/domains/a.example.com/status", serviceAuth, map[string]any{"state": "ready"})
		require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	})

	t.Run("other errors are 400", func(t *testing.T) {
		api, svc := newServiceDomainTestAPI(t)
		svc.ShelfService.EXPECT().SetDomainStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errBoom{})

		require.Equal(t, http.StatusBadRequest, api.Put("/v1/service/domains/a.example.com/status", serviceAuth, body).Code)
	})
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }

func Test_API_ServiceDomains_RejectUserTokens(t *testing.T) {
	api, _ := newServiceDomainTestAPI(t)
	token := signTestJWT(t, "admin-1", model.RoleAdmin, 300_000_000_000)

	require.Equal(t, http.StatusUnauthorized, api.Get("/v1/service/domains", "Authorization: Bearer "+token).Code)
	require.Equal(t, http.StatusUnauthorized, api.Put("/v1/service/domains/a.example.com/status", "Authorization: Bearer "+token, map[string]any{"shelf_id": "s", "state": "ready"}).Code)
}
