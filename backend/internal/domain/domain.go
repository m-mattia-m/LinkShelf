package domain

import (
	"backend/internal/infrastructure/mailer"
	"backend/internal/infrastructure/oidcclient"
	"backend/internal/infrastructure/repository"
)

type Service struct {
	UserService              UserService
	ShelfService             ShelfService
	SectionService           SectionService
	LinkService              LinkService
	SettingService           SettingService
	StatisticService         StatisticService
	AuthService              AuthService
	EmailVerificationService EmailVerificationService
}

// NewService wires up every domain service. oidc is nil when
// authentication.type is not OIDC. m is nil when
// authentication.emailVerification.enabled is false.
func NewService(repository *repository.Repository, oidc *oidcclient.Client, m mailer.Mailer) *Service {
	service := Service{}
	service.UserService = NewUserService(repository, &service)
	service.ShelfService = NewShelfService(repository, &service)
	service.SectionService = NewSectionService(repository, &service)
	service.LinkService = NewLinkService(repository, &service)
	service.SettingService = NewSettingService(repository, &service)
	service.StatisticService = NewStatisticService(repository, &service)
	service.AuthService = NewAuthService(repository, &service, oidc)
	service.EmailVerificationService = NewEmailVerificationService(repository, &service, m)

	return &service
}
