package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/repository"
	"fmt"
	"strings"
)

// BackfillUsernames gives every user that has no username one derived from
// their email address. It runs on every startup and only touches rows that are
// still NULL, so in practice it does real work once: for accounts that existed
// before usernames did. A username is collected for every account whether or
// not app.userBasedPaths is on, and nothing else fills it in for those.
func BackfillUsernames(repo *repository.Repository) error {
	users, err := repo.UserRepository.ListWithoutUsername()
	if err != nil {
		return err
	}

	for _, user := range users {
		username, err := availableUsername(repo, emailLocalPart(user.Email))
		if err != nil {
			return err
		}
		if err := repo.UserRepository.SetUsername(user.Id, username); err != nil {
			return err
		}
	}

	return nil
}

// EnsureUniquePaths refuses to start when app.userBasedPaths is off but two
// shelves share a path. That state only arises by switching the setting off
// after it was on: /<path> would then be ambiguous, so one of the shelves
// could never be reached. The database can't prevent it (its constraint is
// per owner, see migration 0008), so it is checked here instead.
func EnsureUniquePaths(repo *repository.Repository) error {
	if config.Bool("app.userBasedPaths") {
		return nil
	}

	collisions, err := repo.ShelfRepository.ListPathCollisions()
	if err != nil {
		return err
	}
	if len(collisions) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("app.userBasedPaths is false, but these shelves share a path, so /<path> would be ambiguous:\n")
	for _, c := range collisions {
		fmt.Fprintf(&b, "  /%s  shelf %s  owner %s (%s)\n", c.Path, c.ShelfId, c.Username, c.UserId)
	}
	b.WriteString("Rename the shelves so every path is unique, or set app.userBasedPaths back to true.")
	return fmt.Errorf("%s", b.String())
}
