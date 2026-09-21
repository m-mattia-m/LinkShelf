package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"strings"
)

// EnsureBootstrapAdmin idempotently creates (or refreshes the password/role
// of, and fills in a missing username for) the config-driven admin account on
// every startup. It's a no-op if no
// bootstrap admin email/password is configured.
func EnsureBootstrapAdmin(repo *repository.Repository) error {
	email := strings.TrimSpace(config.String("authentication.bootstrapAdmin.email"))
	password := config.String("authentication.bootstrapAdmin.password")
	if email == "" || password == "" {
		return nil
	}

	// Used exactly as configured: the operator chooses this name, so it skips
	// the format and reserved-word checks a user-chosen username goes through.
	// That is what lets it be "admin".
	username := strings.TrimSpace(config.String("authentication.bootstrapAdmin.username"))

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	existing, err := repo.UserRepository.FindByEmail(email)
	if err != nil {
		return err
	}

	// The bootstrap admin is always exempt from email verification - it's
	// the one account an operator needs to be able to log in with
	// immediately, including before SMTP is configured or reachable.
	if existing != nil {
		if existing.Username == "" && username != "" {
			if err := repo.UserRepository.SetUsername(existing.Id, username); err != nil {
				return err
			}
		}
		if err := repo.UserRepository.SetPasswordAndRole(existing.Id, hashedPassword, model.RoleAdmin); err != nil {
			return err
		}
		return repo.UserRepository.MarkVerified(existing.Id)
	}

	userId, err := repo.UserRepository.Create(model.UserBase{
		Email:     email,
		Username:  username,
		FirstName: "Admin",
		LastName:  "Admin",
	}, hashedPassword, model.RoleAdmin)
	if err != nil {
		return err
	}

	if err := repo.UserRepository.SetPasswordAndRole(userId, hashedPassword, model.RoleAdmin); err != nil {
		return err
	}
	return repo.UserRepository.MarkVerified(userId)
}
