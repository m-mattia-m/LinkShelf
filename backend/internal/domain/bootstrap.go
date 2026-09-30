package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"strings"

	"go.uber.org/zap"
)

// EnsureBootstrapAdmin creates the config-driven admin account on startup if
// no account with that email exists yet. It's a no-op if no bootstrap admin
// email/password is configured.
//
// An existing account is never touched: resetting its password and role on
// every restart would silently undo a password changed after the first
// login, and would hand admin back to whoever knows the configured value.
func EnsureBootstrapAdmin(repo *repository.Repository) error {
	email := normalizeEmail(config.String("authentication.bootstrapAdmin.email"))
	password := config.String("authentication.bootstrapAdmin.password")
	if email == "" || password == "" {
		return nil
	}

	if !config.LocalAuthEnabled() {
		zap.L().Warn("authentication.bootstrapAdmin is set, but authentication.localAuthEnabled is false, so it can't sign in with its password")
	}

	existing, err := repo.UserRepository.FindByEmail(email)
	if err != nil {
		return err
	}
	if existing != nil {
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

	userId, err := repo.UserRepository.Create(model.UserBase{
		Email:     email,
		Username:  username,
		FirstName: "Admin",
		LastName:  "Admin",
	}, hashedPassword, model.RoleAdmin)
	if err != nil {
		return err
	}

	// The bootstrap admin is always exempt from email verification - it's
	// the one account an operator needs to be able to log in with
	// immediately, including before SMTP is configured or reachable.
	return repo.UserRepository.MarkVerified(userId)
}
