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
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func enableDomainScopes(t *testing.T) {
	t.Helper()
	enableServiceToken(t)
	config.Set("authentication.serviceToken.scopes", []string{config.ScopeDomainsRead, config.ScopeDomainsWriteStatus})
	t.Cleanup(func() { config.Set("authentication.serviceToken.scopes", []string{config.ScopeLimits}) })
}

func requireReadyConfig(t *testing.T, enabled bool) {
	t.Helper()
	config.Set("customDomains.requireReady", enabled)
	t.Cleanup(func() { config.Set("customDomains.requireReady", false) })
}

func freshDomain() string {
	return "d-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12] + ".example.com"
}

func createDomainShelf(t *testing.T, token, domain string) model.Shelf {
	t.Helper()
	resp := doAuthedRequest(t, http.MethodPost, "/v1/shelves", strings.NewReader(ObjectToJSON(model.ShelfBase{Title: "Domain shelf", Domain: domain})), token)
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(raw))
	var shelf model.Shelf
	require.NoError(t, json.Unmarshal(raw, &shelf))
	return shelf
}

// findInService walks every page of the service list for domain.
func findInService(t *testing.T, domain string) (items []model.ServiceDomain) {
	t.Helper()
	cursor := ""
	for {
		resp := doAuthedRequest(t, http.MethodGet, "/v1/service/domains?limit=2&cursor="+cursor, nil, integrationServiceToken)
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, string(raw))

		var page model.ServiceDomainList
		require.NoError(t, json.Unmarshal(raw, &page))
		for _, item := range page.Items {
			if item.Domain == domain {
				items = append(items, item)
			}
		}
		if page.NextCursor == nil {
			return items
		}
		cursor = *page.NextCursor
	}
}

func putStatus(t *testing.T, domain string, body map[string]any) *http.Response {
	t.Helper()
	return doAuthedRequest(t, http.MethodPut, "/v1/service/domains/"+domain+"/status", strings.NewReader(ObjectToJSON(body)), integrationServiceToken)
}

func publicByDomain(t *testing.T, domain string) int {
	t.Helper()
	resp := doRequest(t, http.MethodGet, "/v1/shelves/by-domain/"+domain, nil)
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

type collectedEvents struct {
	mu     sync.Mutex
	events []string
}

func (c *collectedEvents) Publish(eventType string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	raw, _ := json.Marshal(data)
	c.events = append(c.events, eventType+" "+string(raw))
}

func Test_API_ServiceDomains_RequireTheirScopeAndNoUserToken(t *testing.T) {
	enableServiceToken(t) // scope limits only
	_, adminToken := createTestAdmin(t)

	resp := doAuthedRequest(t, http.MethodGet, "/v1/service/domains", nil, integrationServiceToken)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "the token lacks domains:read")

	resp = doAuthedRequest(t, http.MethodGet, "/v1/service/domains", nil, adminToken)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "not even an admin reaches a service route")

	resp = doRequest(t, http.MethodGet, "/v1/service/domains", nil)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_API_ServiceDomains_TokenIsRejectedOnUserRoutes(t *testing.T) {
	enableDomainScopes(t)

	for _, path := range []string{"/v1/shelves", "/v1/users/me", "/v1/users"} {
		resp := doAuthedRequest(t, http.MethodGet, path, nil, integrationServiceToken)
		_ = resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode, path)
	}
}

func Test_API_ServiceDomains_ListStatusAndEvents(t *testing.T) {
	enableDomainScopes(t)
	events := &collectedEvents{}
	TestService.Events = events
	t.Cleanup(func() { TestService.Events = nil })

	userId, token := createTestUser(t)
	domain := freshDomain()
	shelf := createDomainShelf(t, token, domain)

	items := findInService(t, domain)
	require.Len(t, items, 1)
	require.Equal(t, shelf.Id, items[0].ShelfId)
	require.Equal(t, userId, items[0].OwnerId)
	require.Equal(t, domain, items[0].Status.ObservedDomain)
	require.Empty(t, items[0].Status.State)
	require.NotEqual(t, "1970-01-01T00:00:00Z", items[0].CreatedAt)

	// A page which did not change answers 304.
	resp := doAuthedRequest(t, http.MethodGet, "/v1/service/domains?limit=1", nil, integrationServiceToken)
	_ = resp.Body.Close()
	etag := resp.Header.Get("ETag")
	require.NotEmpty(t, etag)
	req, _ := http.NewRequest(http.MethodGet, BaseURL+"/v1/service/domains?limit=1", nil)
	req.Header.Set("Authorization", "Bearer "+integrationServiceToken)
	req.Header.Set("If-None-Match", etag)
	cached, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = cached.Body.Close()
	require.Equal(t, http.StatusNotModified, cached.StatusCode)

	// The controller writes a status, which the owner then sees on the shelf.
	status := map[string]any{
		"shelf_id": shelf.Id, "state": "pending", "reason": "txt_missing",
		"records": []map[string]string{{"type": "TXT", "name": "_linkshelf-challenge." + domain, "value": "linkshelf-verification=abc"}},
	}
	resp = putStatus(t, domain, status)
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, string(raw))

	own := doAuthedRequest(t, http.MethodGet, "/v1/shelves/"+shelf.Id, nil, token)
	defer func() { _ = own.Body.Close() }()
	var fetched model.Shelf
	ownRaw, _ := io.ReadAll(own.Body)
	require.NoError(t, json.Unmarshal(ownRaw, &fetched))
	require.NotNil(t, fetched.DomainStatus)
	require.Equal(t, "pending", fetched.DomainStatus.State)
	require.Equal(t, "txt_missing", fetched.DomainStatus.Reason)
	require.Len(t, fetched.DomainStatus.Records, 1)

	// A stale controller (the shelf has another domain by now) gets a 409, and the
	// status of the new domain stays empty.
	newDomain := freshDomain()
	update := doAuthedRequest(t, http.MethodPut, "/v1/shelves/"+shelf.Id, strings.NewReader(ObjectToJSON(model.ShelfBase{Title: "Domain shelf", Domain: newDomain})), token)
	_ = update.Body.Close()
	require.Equal(t, http.StatusOK, update.StatusCode)

	resp = putStatus(t, domain, map[string]any{"shelf_id": shelf.Id, "state": "ready"})
	_ = resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode)

	moved := findInService(t, newDomain)
	require.Len(t, moved, 1)
	require.Empty(t, moved[0].Status.State, "changing the domain resets the status")
	require.Empty(t, findInService(t, domain))

	require.Equal(t, http.StatusUnprocessableEntity, func() int {
		r := putStatus(t, newDomain, map[string]any{"shelf_id": shelf.Id, "state": "bogus"})
		_ = r.Body.Close()
		return r.StatusCode
	}())

	// Clearing the domain, and deleting the shelf, are told too.
	del := doAuthedRequest(t, http.MethodDelete, "/v1/shelves/"+shelf.Id, nil, token)
	_ = del.Body.Close()
	require.Equal(t, http.StatusNoContent, del.StatusCode)

	events.mu.Lock()
	defer events.mu.Unlock()
	require.Len(t, events.events, 3)
	require.Contains(t, events.events[0], `"domain":"`+domain+`","previous_domain":""`)
	require.Contains(t, events.events[1], fmt.Sprintf(`"domain":"%s","previous_domain":"%s"`, newDomain, domain))
	require.Contains(t, events.events[2], `"domain":"","previous_domain":"`+newDomain+`"`)
	require.True(t, strings.HasPrefix(events.events[0], "shelf.domain.changed "))
}

func Test_API_ServiceDomains_RequireReadyGatesServingAndUniqueness(t *testing.T) {
	enableDomainScopes(t)
	requireReadyConfig(t, true)

	_, tokenA := createTestUser(t)
	_, tokenB := createTestUser(t)
	domain := freshDomain()

	// Two users can both save the domain, nobody is served yet.
	shelfA := createDomainShelf(t, tokenA, domain)
	shelfB := createDomainShelf(t, tokenB, domain)
	require.Equal(t, http.StatusNotFound, publicByDomain(t, domain))
	require.Len(t, findInService(t, domain), 2, "the controller sees both claims")

	// A pending or failed domain is not served either.
	resp := putStatus(t, domain, map[string]any{"shelf_id": shelfA.Id, "state": "pending"})
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, http.StatusNotFound, publicByDomain(t, domain))

	// The proven owner is served ...
	resp = putStatus(t, domain, map[string]any{"shelf_id": shelfA.Id, "state": "ready"})
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	public := doRequest(t, http.MethodGet, "/v1/shelves/by-domain/"+domain, nil)
	defer func() { _ = public.Body.Close() }()
	require.Equal(t, http.StatusOK, public.StatusCode)
	var served model.PublicShelf
	raw, _ := io.ReadAll(public.Body)
	require.NoError(t, json.Unmarshal(raw, &served))
	require.Equal(t, shelfA.Id, served.Id)

	// ... and the other claim can not become ready, only error out.
	resp = putStatus(t, domain, map[string]any{"shelf_id": shelfB.Id, "state": "ready"})
	_ = resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
	resp = putStatus(t, domain, map[string]any{"shelf_id": shelfB.Id, "state": "error", "reason": "taken"})
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Moving the ready shelf away stops serving at once.
	update := doAuthedRequest(t, http.MethodPut, "/v1/shelves/"+shelfA.Id, strings.NewReader(ObjectToJSON(model.ShelfBase{Title: "Domain shelf", Domain: freshDomain()})), tokenA)
	_ = update.Body.Close()
	require.Equal(t, http.StatusOK, update.StatusCode)
	require.Equal(t, http.StatusNotFound, publicByDomain(t, domain))
}

func Test_API_ServiceDomains_WithoutRequireReadyEverySavedDomainIsServedAndUnique(t *testing.T) {
	enableDomainScopes(t)
	requireReadyConfig(t, false)

	_, tokenA := createTestUser(t)
	_, tokenB := createTestUser(t)
	domain := freshDomain()

	createDomainShelf(t, tokenA, domain)
	require.Equal(t, http.StatusOK, publicByDomain(t, domain), "status is informational")

	resp := doAuthedRequest(t, http.MethodPost, "/v1/shelves", strings.NewReader(ObjectToJSON(model.ShelfBase{Title: "Second", Domain: domain})), tokenB)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode, "one domain, one shelf, as before")
}
