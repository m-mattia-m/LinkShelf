package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/repository"
	"fmt"
	"regexp"
	"strings"
)

const (
	usernameMinLength = 3
	usernameMaxLength = 30
	// usernameSuffixLimit bounds the "-2", "-3", ... search for a free name.
	usernameSuffixLimit = 10000
)

// usernamePattern: lowercase letters, digits and inner hyphens.
var usernamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// routeReservedNames are top-level routes of the frontend or backend.
var routeReservedNames = map[string]struct{}{
	"app": {}, "auth": {}, "docs": {}, "cloud": {}, "about": {}, "contact": {},
	"imprint": {}, "privacy-policy": {}, "terms-of-use": {},
	"api": {}, "v1": {}, "swagger": {}, "health": {}, "images": {},
}

// otherReservedNames are blocked as usernames only.
var otherReservedNames = map[string]struct{}{
	"admin": {}, "root": {}, "support": {}, "help": {}, "static": {}, "assets": {},
	"public": {}, "login": {}, "logout": {}, "sign-in": {}, "sign-up": {},
	"register": {}, "settings": {}, "dashboard": {}, "profile": {}, "user": {},
	"users": {}, "me": {}, "www": {}, "null": {}, "undefined": {},
}

// isRouteReserved reports whether name is a route, including the assets base path.
func isRouteReserved(name string) bool {
	name = strings.ToLower(name)
	if _, ok := routeReservedNames[name]; ok {
		return true
	}

	assetsBase := strings.Trim(config.String("assets.basePath"), "/")
	if first, _, _ := strings.Cut(assetsBase, "/"); first != "" && strings.ToLower(first) == name {
		return true
	}
	return false
}

func isReservedUsername(name string) bool {
	if isRouteReserved(name) {
		return true
	}
	_, ok := otherReservedNames[strings.ToLower(name)]
	return ok
}

// validateUsername checks format and reserved words, not availability.
func validateUsername(username string) error {
	if len(username) < usernameMinLength || len(username) > usernameMaxLength || !usernamePattern.MatchString(username) {
		return fmt.Errorf("%w: username must be %d-%d characters of lowercase letters, numbers and hyphens, and must not start or end with a hyphen",
			ErrInvalidInput, usernameMinLength, usernameMaxLength)
	}
	if isReservedUsername(username) {
		return fmt.Errorf("%w: username %q is reserved", ErrInvalidInput, username)
	}
	return nil
}

// checkUsernameAvailable returns ErrConflict if another user has the username.
func checkUsernameAvailable(repo *repository.Repository, username, exceptUserId string) error {
	taken, err := repo.UserRepository.UsernameTaken(username, exceptUserId)
	if err != nil {
		return err
	}
	if taken {
		return fmt.Errorf("%w: username %q is already taken", ErrConflict, username)
	}
	return nil
}

// sanitizeUsername maps arbitrary text to the username alphabet. The result may be invalid.
func sanitizeUsername(raw string) string {
	var b strings.Builder
	lastHyphen := true // swallows leading hyphens
	for _, r := range strings.ToLower(raw) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastHyphen = false
			continue
		}
		if !lastHyphen {
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// emailLocalPart is everything before the last "@" - "myname@example.com" -> "myname".
func emailLocalPart(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return email
	}
	return email[:at]
}

// availableUsername derives an unused username, appending -2, -3, ... if needed.
func availableUsername(repo *repository.Repository, candidates ...string) (string, error) {
	base := "member"
	for _, candidate := range candidates {
		if sanitized := sanitizeUsername(candidate); sanitized != "" {
			base = sanitized
			break
		}
	}

	for n := 1; n <= usernameSuffixLimit; n++ {
		candidate := base
		if n > 1 {
			suffix := fmt.Sprintf("-%d", n)
			trimmed := strings.TrimRight(truncate(base, usernameMaxLength-len(suffix)), "-")
			candidate = trimmed + suffix
		} else {
			candidate = strings.TrimRight(truncate(candidate, usernameMaxLength), "-")
		}

		if validateUsername(candidate) != nil {
			continue
		}
		taken, err := repo.UserRepository.UsernameTaken(candidate, "")
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not find a free username based on %q", base)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
