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

// app.strictOrigins: originPolicy handles CORS, hostGuard checks the request host.

const (
	// domainCacheTtl is how long a shelf-domain lookup is cached.
	domainCacheTtl = 60 * time.Second
	// domainCacheMaxEntries bounds the cache, since the Origin header is client-controlled.
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
	// additionalOrigins are other origins of the instance itself.
	additionalOrigins map[string]bool
	// allowHttpDomains allows http shelf domains when the frontend itself runs on http.
	allowHttpDomains bool
	registered       func(domain string) (bool, error)
	cache            *domainCache
}

func newOriginPolicy(svc *domain.Service) *originPolicy {
	frontend, _ := url.Parse(strings.TrimSpace(config.String("app.frontendUrl")))

	additional := make(map[string]bool)
	for _, origin := range config.Strings("app.additionalOrigins") {
		parsed, _ := url.Parse(strings.TrimSpace(origin))
		if canonical := canonicalOrigin(parsed); canonical != "" {
			additional[canonical] = true
		}
	}

	return &originPolicy{
		frontendOrigin:    canonicalOrigin(frontend),
		additionalOrigins: additional,
		allowHttpDomains:  frontend != nil && frontend.Scheme == "http",
		registered:        svc.ShelfService.IsDomainRegistered,
		cache:             newDomainCache(),
	}
}

// canonicalOrigin returns scheme://host[:port] as browsers send it, or "".
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
	if canonical == p.frontendOrigin || p.additionalOrigins[canonical] {
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

// proxyMatcher reports whether the peer is a trusted proxy (for X-Forwarded-Host).
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

// hostMatches reports whether requestHost is the configured server.host (and port).
func hostMatches(requestHost string) bool {
	wantHost := strings.TrimRight(strings.ToLower(strings.TrimSpace(config.String("server.host"))), ".")
	got := domain.NormalizeDomain(requestHost)

	if config.Bool("domain.openapi.usePort") {
		return got == domain.NormalizeDomain(wantHost+":"+strconv.Itoa(config.Int("server.port")))
	}

	parsed, err := url.Parse("//" + got)
	return err == nil && parsed.Hostname() == wantHost
}

// hostGuard answers 421 to requests not addressed to server.host, except health checks.
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
