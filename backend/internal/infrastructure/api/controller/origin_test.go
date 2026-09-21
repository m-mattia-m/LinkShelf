package controller

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testOriginPolicy(frontendUrl string, registered map[string]bool) (*originPolicy, *int) {
	parsed, _ := url.Parse(frontendUrl)
	calls := 0
	return &originPolicy{
		frontendOrigin:   canonicalOrigin(parsed),
		allowHttpDomains: parsed.Scheme == "http",
		registered: func(domain string) (bool, error) {
			calls++
			return registered[domain], nil
		},
		cache: newDomainCache(),
	}, &calls
}

func Test_Unit_CanonicalOrigin(t *testing.T) {
	cases := map[string]string{
		"https://Linkshelf.Example.com":      "https://linkshelf.example.com",
		"https://linkshelf.example.com:443":  "https://linkshelf.example.com",
		"http://linkshelf.example.com:80":    "http://linkshelf.example.com",
		"https://linkshelf.example.com:9443": "https://linkshelf.example.com:9443",
		"http://linkshelf.example.com:443":   "http://linkshelf.example.com:443",
		"https://linkshelf.example.com/":     "https://linkshelf.example.com",
		"http://localhost:3000":              "http://localhost:3000",
		"ftp://linkshelf.example.com":        "",
		"linkshelf.example.com":              "",
		"null":                               "",
		"":                                   "",
	}

	for input, want := range cases {
		parsed, _ := url.Parse(input)
		require.Equal(t, want, canonicalOrigin(parsed), input)
	}
	require.Empty(t, canonicalOrigin(nil))
}

func Test_Unit_OriginPolicy_Allowed(t *testing.T) {
	registered := map[string]bool{"profile.example.com": true, "profile.example.com:9443": true}

	t.Run("the frontend's own origin", func(t *testing.T) {
		policy, calls := testOriginPolicy("https://linkshelf.example.com", registered)

		require.True(t, policy.allowed("https://linkshelf.example.com"))
		require.True(t, policy.allowed("https://LinkShelf.example.com:443"))
		require.Zero(t, *calls, "the frontend origin needs no database lookup")
	})

	t.Run("a registered shelf domain over https", func(t *testing.T) {
		policy, _ := testOriginPolicy("https://linkshelf.example.com", registered)

		require.True(t, policy.allowed("https://profile.example.com"))
		require.True(t, policy.allowed("https://profile.example.com:9443"))
	})

	t.Run("a domain no shelf has", func(t *testing.T) {
		policy, _ := testOriginPolicy("https://linkshelf.example.com", registered)

		require.False(t, policy.allowed("https://evil.example.org"))
		require.False(t, policy.allowed("https://profile.example.com:8443"), "another port is another site")
	})

	t.Run("http shelf domains only when the frontend itself is on http", func(t *testing.T) {
		policy, _ := testOriginPolicy("https://linkshelf.example.com", registered)
		require.False(t, policy.allowed("http://profile.example.com"))

		policy, _ = testOriginPolicy("http://linkshelf.example.com", registered)
		require.True(t, policy.allowed("http://profile.example.com"))
		require.True(t, policy.allowed("https://profile.example.com"))
	})

	t.Run("things that aren't an origin", func(t *testing.T) {
		policy, calls := testOriginPolicy("https://linkshelf.example.com", registered)

		for _, origin := range []string{"", "null", "file://", "chrome-extension://abc", "https://profile.example.com/path", "https://localhost", "https://1.2.3.4", "://", "https://%zz"} {
			require.False(t, policy.allowed(origin), origin)
		}
		require.Zero(t, *calls, "nothing that can't be a domain reaches the database")
	})

	t.Run("a failing lookup denies and isn't remembered", func(t *testing.T) {
		policy, _ := testOriginPolicy("https://linkshelf.example.com", registered)
		fail := true
		policy.registered = func(string) (bool, error) {
			if fail {
				return false, errors.New("db unavailable")
			}
			return true, nil
		}

		require.False(t, policy.allowed("https://profile.example.com"))
		fail = false
		require.True(t, policy.allowed("https://profile.example.com"))
	})
}

func Test_Unit_DomainCache(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	cache := newDomainCache()
	cache.now = func() time.Time { return now }

	_, found := cache.get("a.example.com")
	require.False(t, found)

	cache.set("a.example.com", true)
	cache.set("b.example.com", false)
	registered, found := cache.get("a.example.com")
	require.True(t, found)
	require.True(t, registered)
	registered, found = cache.get("b.example.com")
	require.True(t, found)
	require.False(t, registered, "a miss is cached too")

	now = now.Add(domainCacheTtl - time.Second)
	_, found = cache.get("a.example.com")
	require.True(t, found)

	now = now.Add(2 * time.Second)
	_, found = cache.get("a.example.com")
	require.False(t, found, "an entry expires after the ttl")
}

func Test_Unit_DomainCache_IsBounded(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	cache := newDomainCache()
	cache.now = func() time.Time { return now }

	for i := 0; i < domainCacheMaxEntries; i++ {
		cache.set("d"+time.Duration(i).String()+".example.com", true)
	}
	require.Len(t, cache.entries, domainCacheMaxEntries)

	// Full of live entries: it starts over instead of growing.
	cache.set("one-more.example.com", true)
	require.LessOrEqual(t, len(cache.entries), domainCacheMaxEntries)
	_, found := cache.get("one-more.example.com")
	require.True(t, found)

	// Full of expired entries: only those are dropped.
	now = now.Add(2 * domainCacheTtl)
	for len(cache.entries) < domainCacheMaxEntries {
		cache.entries["fill"+time.Duration(len(cache.entries)).String()] = domainCacheEntry{expires: now.Add(-time.Second)}
	}
	cache.set("fresh.example.com", true)
	require.Len(t, cache.entries, 1)
}

func Test_Unit_ProxyMatcher(t *testing.T) {
	matcher := newProxyMatcher([]string{"127.0.0.1", "10.0.0.0/8", " ::1 ", "not-an-ip"})

	for _, ip := range []string{"127.0.0.1", "10.1.2.3", "::1", "::ffff:127.0.0.1"} {
		require.True(t, matcher.trusts(ip), ip)
	}
	for _, ip := range []string{"192.168.0.1", "11.0.0.1", "", "junk"} {
		require.False(t, matcher.trusts(ip), ip)
	}
	require.False(t, newProxyMatcher(nil).trusts("127.0.0.1"))
}

func Test_Unit_HostMatches(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	config.Set("server.host", "API.Example.com")
	config.Set("server.port", 8085)

	t.Run("without usePort only the host counts", func(t *testing.T) {
		config.Set("domain.openapi.usePort", false)

		for _, host := range []string{"api.example.com", "API.example.com", "api.example.com:8085", "api.example.com:9999", "api.example.com.", "api.example.com:443"} {
			require.True(t, hostMatches(host), host)
		}
		for _, host := range []string{"", "example.com", "evil.example.org", "api.example.com.evil.org", "10.0.0.5:8085", "xapi.example.com"} {
			require.False(t, hostMatches(host), host)
		}
	})

	t.Run("with usePort the port has to match as well", func(t *testing.T) {
		config.Set("domain.openapi.usePort", true)

		require.True(t, hostMatches("api.example.com:8085"))
		require.False(t, hostMatches("api.example.com"))
		require.False(t, hostMatches("api.example.com:9999"))
	})

	t.Run("with usePort a default port is left out of the header", func(t *testing.T) {
		config.Set("domain.openapi.usePort", true)
		config.Set("server.port", 443)

		require.True(t, hostMatches("api.example.com"))
		require.True(t, hostMatches("api.example.com:443"))
		require.False(t, hostMatches("api.example.com:8085"))
	})
}

// strictRouter builds the real router with app.strictOrigins on, so the
// middleware order (host guard, then CORS) is part of what is tested.
func strictRouter(t *testing.T, svc *MockService) http.Handler {
	t.Helper()

	config.Reset()
	t.Cleanup(config.Reset)
	require.NoError(t, config.LoadConfig())
	config.Set("app.strictOrigins", true)
	config.Set("app.frontendUrl", "https://linkshelf.example.com")
	config.Set("server.host", "api.example.com")
	config.Set("domain.openapi.usePort", false)
	config.Set("server.trustedProxies", []string{"127.0.0.1"})

	router, err := Router(svc.Service)
	require.NoError(t, err)
	return router
}

func strictRequest(host, origin string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/v1/shelves/by-domain/profile.example.com", nil)
	req.Host = host
	req.RemoteAddr = "192.168.1.5:4242"
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	return req
}

func Test_Router_StrictOrigins_HostGuard(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()
	svc.ShelfService.EXPECT().
		GetByDomain("profile.example.com").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-1"}}, nil).
		AnyTimes()
	router := strictRouter(t, svc)

	t.Run("the configured host is served", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, strictRequest("api.example.com", ""))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("another host gets a bare 421", func(t *testing.T) {
		for _, host := range []string{"10.0.0.5:8085", "linkshelf.example.com", "evil.example.org"} {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, strictRequest(host, ""))
			require.Equal(t, http.StatusMisdirectedRequest, rec.Code, host)
			require.Empty(t, rec.Body.String(), host)
		}
	})

	t.Run("the health endpoints answer on any host", func(t *testing.T) {
		for _, path := range []string{"/health/liveness", "/health/readiness"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Host = "10.42.0.7:8085"
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code, path)
		}
	})

	t.Run("X-Forwarded-Host counts only from a trusted proxy", func(t *testing.T) {
		req := strictRequest("10.0.0.5:8085", "")
		req.Header.Set("X-Forwarded-Host", "api.example.com")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMisdirectedRequest, rec.Code, "an untrusted peer can't vouch for the host")

		req = strictRequest("10.0.0.5:8085", "")
		req.RemoteAddr = "127.0.0.1:4242"
		req.Header.Set("X-Forwarded-Host", "api.example.com, other.example.com")
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, "the first forwarded host is the one that was asked for")

		req = strictRequest("api.example.com", "")
		req.RemoteAddr = "127.0.0.1:4242"
		req.Header.Set("X-Forwarded-Host", "evil.example.org")
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMisdirectedRequest, rec.Code, "the forwarded host replaces the Host header")
	})
}

func Test_Router_StrictOrigins_Cors(t *testing.T) {
	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()
	svc.ShelfService.EXPECT().
		GetByDomain("profile.example.com").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-1"}}, nil).
		AnyTimes()
	svc.ShelfService.EXPECT().IsDomainRegistered("profile.example.com").Return(true, nil).MaxTimes(1)
	svc.ShelfService.EXPECT().IsDomainRegistered("evil.example.org").Return(false, nil).MaxTimes(1)
	router := strictRouter(t, svc)

	t.Run("the frontend origin may call the API", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, strictRequest("api.example.com", "https://linkshelf.example.com"))
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "https://linkshelf.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("a shelf domain may call the API", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, strictRequest("api.example.com", "https://profile.example.com"))
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "https://profile.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("any other origin is refused", func(t *testing.T) {
		for range 2 { // the second time it comes from the cache
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, strictRequest("api.example.com", "https://evil.example.org"))
			require.Equal(t, http.StatusForbidden, rec.Code)
			require.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("a preflight from an allowed origin is answered", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/v1/links", nil)
		req.Host = "api.example.com"
		req.Header.Set("Origin", "https://linkshelf.example.com")
		req.Header.Set("Access-Control-Request-Method", "PUT")
		req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
		require.Equal(t, "https://linkshelf.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
		require.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("no Origin header means it isn't a browser CORS request", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, strictRequest("api.example.com", ""))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("the API's own origin, such as its swagger page, is not a cross-origin call", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, strictRequest("api.example.com", "https://api.example.com"))
		require.Equal(t, http.StatusOK, rec.Code)
	})
}

func Test_Router_WithoutStrictOrigins_AcceptsEverything(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	require.NoError(t, config.LoadConfig())
	require.False(t, config.Bool("app.strictOrigins"), "the shipped default must stay open")

	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()
	svc.ShelfService.EXPECT().
		GetByDomain("profile.example.com").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "shelf-1"}}, nil)
	router, err := Router(svc.Service)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, strictRequest("whatever.example.net", "https://evil.example.org"))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func Test_Router_ServesThePublicShelfByDomain(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	require.NoError(t, config.LoadConfig())

	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()
	svc.ShelfService.EXPECT().
		GetByDomain("profile.example.com:9443").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "by-domain", Title: "Profile"}, UserId: "secret-owner"}, nil)
	svc.ShelfService.EXPECT().GetByDomain("nobody.example.com").Return(nil, nil)
	svc.ShelfService.EXPECT().GetByDomain("broken.example.com").Return(nil, errors.New("db unavailable"))
	router, err := Router(svc.Service)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/shelves/by-domain/profile.example.com%3A9443", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"id":"by-domain"`)
	require.NotContains(t, rec.Body.String(), "secret-owner", "only the public fields are exposed")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/shelves/by-domain/nobody.example.com", nil))
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/shelves/by-domain/broken.example.com", nil))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
