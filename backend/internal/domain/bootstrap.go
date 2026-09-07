package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"strings"
)

// EnsureBootstrapAdmin idempotently creates (or refreshes the password/role
// of) the config-driven admin account on every startup. It's a no-op if no
// bootstrap admin email/password is configured.
func EnsureBootstrapAdmin(repo *repository.Repository) error {
	email := strings.TrimSpace(config.String("authentication.bootstrapAdmin.email"))
	password := config.String("authentication.bootstrapAdmin.password")
	if email == "" || password == "" {
		return nil
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	existing, err := repo.UserRepository.FindByEmail(email)
	if err != nil {
		return err
	}

	if existing != nil {
		return repo.UserRepository.SetPasswordAndRole(existing.Id, hashedPassword, model.RoleAdmin)
	}

	userId, err := repo.UserRepository.Create(model.UserBase{
		Email:     email,
		FirstName: "Admin",
		LastName:  "Admin",
	}, hashedPassword, model.RoleAdmin)
	if err != nil {
		return err
	}

	return repo.UserRepository.SetPasswordAndRole(userId, hashedPassword, model.RoleAdmin)
}
