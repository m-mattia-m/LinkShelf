package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/repository"
	"fmt"
	"strings"
)

// BackfillUsernames derives a username for users that have none.
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

// EnsureUniquePaths refuses to start if shelves share a path while app.userBasedPaths is off.
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
