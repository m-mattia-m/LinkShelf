package domain

import "backend/internal/infrastructure/repository"

type Service struct {
	UserService    UserService
	ShelfService   ShelfService
	SectionService SectionService
	LinkService    LinkService
	SettingService SettingService
}

func NewService(repository *repository.Repository) *Service {
	service := Service{}
	service.UserService = NewUserService(repository, &service)
	service.ShelfService = NewShelfService(repository, &service)
	service.SectionService = NewSectionService(repository, &service)
	service.LinkService = NewLinkService(repository, &service)
	service.SettingService = NewSettingService(repository, &service)

	return &service
}
