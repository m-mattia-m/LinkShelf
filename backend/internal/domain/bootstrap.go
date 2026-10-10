package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"strings"

	"go.uber.org/zap"
)

// EnsureBootstrapAdmin creates the configured admin account if it doesn't exist yet.
// An existing account is never modified.
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

	// The configured username skips validation, so it may be "admin".
	username := strings.TrimSpace(config.String("authentication.bootstrapAdmin.username"))

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	defaultMax, err := config.DefaultMaxShelves()
	if err != nil {
		return err
	}

	userId, err := repo.UserRepository.Create(model.UserBase{
		Email:     email,
		Username:  username,
		FirstName: "Admin",
		LastName:  "Admin",
	}, hashedPassword, model.RoleAdmin, defaultMax)
	if err != nil {
		return err
	}

	// The bootstrap admin is exempt from email verification.
	return repo.UserRepository.MarkVerified(userId)
}
