package domain

import (
	"backend/internal/config"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	// domainMaxLength is the longest a DNS name may be (RFC 1035), without the
	// optional port.
	domainMaxLength = 253
	// domainLabelMaxLength is the longest a single DNS label may be.
	domainLabelMaxLength = 63
	maxPort              = 65535
)

var (
	// domainLabelPattern is one DNS label: letters, digits and hyphens, not
	// starting or ending with a hyphen (RFC 1123).
	domainLabelPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	// domainTldPattern is the last label: letters only (so IP addresses never
	// pass), or a punycode-encoded international TLD.
	domainTldPattern = regexp.MustCompile(`^([a-z]{2,}|xn--[a-z0-9]([a-z0-9-]*[a-z0-9])?)$`)
	// domainPortPattern is a port without leading zeros, so "0443" can't be
	// used to dodge the default-port stripping in NormalizeDomain.
	domainPortPattern = regexp.MustCompile(`^[1-9][0-9]{0,4}$`)
)

// NormalizeDomain trims and lowercases a domain and strips a trailing slash or
// dot and the default ports :80 and :443. It never rejects anything; see
// ValidateDomain. frontend/shared/utils/shelfDomain.ts implements the same
// rules, keep them in step.
func NormalizeDomain(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.TrimRight(s, "/")

	host, port, hasPort := s, "", false
	if i := strings.LastIndex(s, ":"); i >= 0 {
		host, port, hasPort = s[:i], s[i+1:], true
	}
	host = strings.TrimRight(host, ".")

	if hasPort && port != "80" && port != "443" {
		return host + ":" + port
	}
	return host
}

// ValidateDomain checks an already normalized domain: a DNS name with a
// letters-only (or punycode) TLD and an optional port.
func ValidateDomain(domain string) error {
	if strings.Contains(domain, "://") || strings.Contains(domain, "/") {
		return fmt.Errorf("%w: a domain must not contain a scheme or a path", ErrInvalidInput)
	}

	host, port, hasPort := domain, "", false
	if i := strings.LastIndex(domain, ":"); i >= 0 {
		host, port, hasPort = domain[:i], domain[i+1:], true
	}

	if hasPort {
		number, err := strconv.Atoi(port)
		if !domainPortPattern.MatchString(port) || err != nil || number > maxPort {
			return fmt.Errorf("%w: %q is not a valid port", ErrInvalidInput, port)
		}
	}

	if host == "" || len(host) > domainMaxLength {
		return fmt.Errorf("%w: a domain must be 1-%d characters", ErrInvalidInput, domainMaxLength)
	}

	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return fmt.Errorf("%w: %q is not a fully qualified domain name (for example profile.example.com)", ErrInvalidInput, host)
	}
	for i, label := range labels {
		if len(label) > domainLabelMaxLength || !domainLabelPattern.MatchString(label) {
			return fmt.Errorf("%w: %q is not a valid domain name", ErrInvalidInput, host)
		}
		if i == len(labels)-1 && !domainTldPattern.MatchString(label) {
			return fmt.Errorf("%w: %q does not end in a valid top-level domain", ErrInvalidInput, host)
		}
	}
	return nil
}

// reservedShelfDomains are the hosts this instance itself is served on, with
// and without port, so a shelf can't take over the app. Unset URLs contribute
// nothing.
func reservedShelfDomains() map[string]struct{} {
	reserved := make(map[string]struct{})
	add := func(host, port string) {
		host = strings.TrimRight(strings.ToLower(strings.TrimSpace(host)), ".")
		if host == "" {
			return
		}
		reserved[host] = struct{}{}
		if port != "" {
			reserved[NormalizeDomain(host+":"+port)] = struct{}{}
		}
	}

	origins := append([]string{config.String("app.frontendUrl")}, config.Strings("app.additionalOrigins")...)
	for _, origin := range origins {
		if parsed, err := url.Parse(strings.TrimSpace(origin)); err == nil {
			add(parsed.Hostname(), parsed.Port())
		}
	}

	serverPort := ""
	if config.Int("server.port") > 0 {
		serverPort = strconv.Itoa(config.Int("server.port"))
	}
	add(config.String("server.host"), serverPort)

	return reserved
}

// checkShelfDomain normalizes and validates domain and rejects this
// instance's own hosts. It does not check whether the domain is taken.
func checkShelfDomain(domain string) (string, error) {
	domain = NormalizeDomain(domain)
	if domain == "" {
		return "", nil
	}
	if err := ValidateDomain(domain); err != nil {
		return "", err
	}
	if _, taken := reservedShelfDomains()[domain]; taken {
		return "", fmt.Errorf("%w: the domain %q belongs to this instance and can't be used for a shelf", ErrInvalidInput, domain)
	}
	return domain, nil
}
