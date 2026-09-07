package domain

import (
	"backend/internal/infrastructure/oidcclient"
	"backend/internal/infrastructure/repository"
)

type Service struct {
	UserService      UserService
	ShelfService     ShelfService
	SectionService   SectionService
	LinkService      LinkService
	SettingService   SettingService
	StatisticService StatisticService
	AuthService      AuthService
}

// NewService wires up every domain service. oidc is nil when
// authentication.type is not OIDC.
func NewService(repository *repository.Repository, oidc *oidcclient.Client) *Service {
	service := Service{}
	service.UserService = NewUserService(repository, &service)
	service.ShelfService = NewShelfService(repository, &service)
	service.SectionService = NewSectionService(repository, &service)
	service.LinkService = NewLinkService(repository, &service)
	service.SettingService = NewSettingService(repository, &service)
	service.StatisticService = NewStatisticService(repository, &service)
	service.AuthService = NewAuthService(repository, &service, oidc)

	return &service
}
