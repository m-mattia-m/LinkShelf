package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// This file implements app.strictOrigins: the API only serves what belongs to
// this instance. Two independent checks make that up.
//
//   - originPolicy decides which browser origins may call the API (CORS): the
//     origin of app.frontendUrl, and the domain of any shelf that is served on
//     one - a shelf page on its own domain loads its content from the API in
//     the visitor's browser, so its origin has to be let through.
//   - hostGuard only answers requests that are addressed to server.host, so
//     the API can't be reached through some other name that happens to point
//     at it.
//
// Neither is authentication: a request without an Origin header (curl, the
// frontend's server-side calls) is not a CORS request and is left alone, and
// every endpoint still checks its own token.

const (
	// domainCacheTtl is how long a "is this a shelf domain?" answer is
	// remembered, so a page load doesn't turn into a database query per API
	// call. A domain that was just added or removed takes up to this long to
	// take effect for CORS.
	domainCacheTtl = 60 * time.Second
	// domainCacheMaxEntries bounds the cache: the Origin header is chosen by
	// whoever sends the request, so without a limit anyone could grow it.
	domainCacheMaxEntries = 1024
)

// domainCache remembers, per domain, whether a shelf is served on it.
type domainCache struct {
	mu      sync.Mutex
	entries map[string]domainCacheEntry
	now     func() time.Time
}

type domainCacheEntry struct {
	registered bool
	expires    time.Time
}

func newDomainCache() *domainCache {
	return &domainCache{entries: make(map[string]domainCacheEntry), now: time.Now}
}

func (c *domainCache) get(domain string) (registered, found bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[domain]
	if !ok || !c.now().Before(entry.expires) {
		return false, false
	}
	return entry.registered, true
}

func (c *domainCache) set(domain string, registered bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= domainCacheMaxEntries {
		now := c.now()
		for key, entry := range c.entries {
			if !now.Before(entry.expires) {
				delete(c.entries, key)
			}
		}
		// Still full of live entries: start over rather than refuse to cache.
		if len(c.entries) >= domainCacheMaxEntries {
			c.entries = make(map[string]domainCacheEntry)
		}
	}
	c.entries[domain] = domainCacheEntry{registered: registered, expires: c.now().Add(domainCacheTtl)}
}

// originPolicy answers whether a browser origin may call the API.
type originPolicy struct {
	frontendOrigin string
	// allowHttpDomains lets a shelf domain be called over plain http too. That
	// is only for development and LAN setups, so it follows the frontend: an
	// instance that runs on http:// itself has nothing to protect by being
	// stricter about its shelves.
	allowHttpDomains bool
	registered       func(domain string) (bool, error)
	cache            *domainCache
}

func newOriginPolicy(svc *domain.Service) *originPolicy {
	frontend, _ := url.Parse(strings.TrimSpace(config.String("app.frontendUrl")))
	return &originPolicy{
		frontendOrigin:   canonicalOrigin(frontend),
		allowHttpDomains: frontend != nil && frontend.Scheme == "http",
		registered:       svc.ShelfService.IsDomainRegistered,
		cache:            newDomainCache(),
	}
}

// canonicalOrigin is scheme://host[:port] in lowercase and without the port
// when it is the scheme's default, which is how a browser writes the Origin
// header. Anything that isn't a plain http(s) origin gives "".
func canonicalOrigin(u *url.URL) string {
	if u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return ""
	}

	host := strings.ToLower(strings.TrimRight(u.Hostname(), "."))
	port := u.Port()
	if (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		host += ":" + port
	}
	return u.Scheme + "://" + host
}

func (p *originPolicy) allowed(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	canonical := canonicalOrigin(parsed)
	if canonical == "" || (parsed.Path != "" && parsed.Path != "/") {
		return false
	}
	if canonical == p.frontendOrigin {
		return true
	}

	if parsed.Scheme == "http" && !p.allowHttpDomains {
		return false
	}

	shelfDomain := domain.NormalizeDomain(parsed.Host)
	if domain.ValidateDomain(shelfDomain) != nil {
		return false
	}

	if registered, found := p.cache.get(shelfDomain); found {
		return registered
	}
	registered, err := p.registered(shelfDomain)
	if err != nil {
		// Fail closed, and don't remember the failure.
		zap.L().Warn("could not check whether an origin is a shelf domain", zap.String("domain", shelfDomain), zap.Error(err))
		return false
	}
	p.cache.set(shelfDomain, registered)
	return registered
}

// proxyMatcher decides whether the peer that opened the connection is one of
// server.trustedProxies, the only ones whose X-Forwarded-Host is believed.
type proxyMatcher []netip.Prefix

func newProxyMatcher(entries []string) proxyMatcher {
	matcher := make(proxyMatcher, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			matcher = append(matcher, prefix)
		} else if addr, err := netip.ParseAddr(entry); err == nil {
			matcher = append(matcher, netip.PrefixFrom(addr, addr.BitLen()))
		}
	}
	return matcher
}

func (m proxyMatcher) trusts(remoteIp string) bool {
	addr, err := netip.ParseAddr(remoteIp)
	if err != nil {
		return false
	}
	for _, prefix := range m {
		if prefix.Contains(addr.Unmap()) {
			return true
		}
	}
	return false
}

// hostMatches reports whether requestHost, as sent by the client, is the host
// the API is configured to be served on: server.host, and server.port too
// when domain.openapi.usePort says the public address includes it.
func hostMatches(requestHost string) bool {
	wantHost := strings.TrimRight(strings.ToLower(strings.TrimSpace(config.String("server.host"))), ".")
	got := domain.NormalizeDomain(requestHost)

	if config.Bool("domain.openapi.usePort") {
		return got == domain.NormalizeDomain(wantHost+":"+strconv.Itoa(config.Int("server.port")))
	}

	parsed, err := url.Parse("//" + got)
	return err == nil && parsed.Hostname() == wantHost
}

// hostGuard answers 421 Misdirected Request to anything not addressed to
// server.host. The health endpoints are exempt: Kubernetes probes address the
// pod by its IP.
func hostGuard() gin.HandlerFunc {
	trusted := newProxyMatcher(config.Strings("server.trustedProxies"))

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/health/") {
			c.Next()
			return
		}

		host := c.Request.Host
		if forwarded := c.GetHeader("X-Forwarded-Host"); forwarded != "" && trusted.trusts(c.RemoteIP()) {
			host = strings.TrimSpace(strings.Split(forwarded, ",")[0])
		}

		if !hostMatches(host) {
			c.AbortWithStatus(http.StatusMisdirectedRequest)
			return
		}
		c.Next()
	}
}
