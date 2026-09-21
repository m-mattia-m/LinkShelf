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

// usernamePattern is the whole username format: lowercase letters, digits and
// hyphens, not starting or ending with a hyphen. Usernames are always stored
// lowercase, which is what lets a plain UNIQUE column act case-insensitively
// on both Postgres and MySQL.
var usernamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// routeReservedNames are the words a top-level URL segment already belongs to:
// the frontend's static pages and the backend's own endpoints. A username
// with one of these names would shadow (or be shadowed by) that route, and
// with user-based paths off a shelf path with one of these names could never
// be reached.
var routeReservedNames = map[string]struct{}{
	"app": {}, "auth": {}, "docs": {}, "cloud": {}, "about": {}, "contact": {},
	"imprint": {}, "privacy-policy": {}, "terms-of-use": {},
	"api": {}, "v1": {}, "swagger": {}, "health": {}, "images": {},
}

// otherReservedNames collide with no route today but are too likely to be
// mistaken for the instance itself (or to become routes), so they can't be
// picked as a username. They are irrelevant for shelf paths.
var otherReservedNames = map[string]struct{}{
	"admin": {}, "root": {}, "support": {}, "help": {}, "static": {}, "assets": {},
	"public": {}, "login": {}, "logout": {}, "sign-in": {}, "sign-up": {},
	"register": {}, "settings": {}, "dashboard": {}, "profile": {}, "user": {},
	"users": {}, "me": {}, "www": {}, "null": {}, "undefined": {},
}

// isRouteReserved reports whether name (compared case-insensitively) is taken
// by a route. The configured assets base path counts too, since that is where
// the backend serves uploaded files from.
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

// validateUsername checks the format and the reserved words. It does not
// check whether the name is taken - that needs the database (see
// checkUsernameAvailable).
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

// checkUsernameAvailable returns ErrConflict when another user already has the
// username. exceptUserId is the user being renamed (empty for a new user).
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

// sanitizeUsername turns arbitrary text (an OIDC claim, an email prefix) into
// the username alphabet: lowercase, every other character becomes a hyphen,
// runs of hyphens collapse and the ends are trimmed. The result may still be
// empty, too short or too long - see availableUsername.
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

// availableUsername derives a valid, unused username from the first candidate
// that yields anything after sanitizing, falling back to "member" ("user" is
// itself reserved). When the
// result is too short, reserved or taken it appends -2, -3, ... until one
// works, shortening the base so the suffix always fits.
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
