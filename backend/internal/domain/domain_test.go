package domain

import (
	"backend/internal/infrastructure/repository"
	"backend/internal/infrastructure/repository/mocks"
	"testing"

	"go.uber.org/mock/gomock"
)

type MockService struct {
	Ctrl    *gomock.Controller
	Service *Service

	UserRepository    *mocks.MockUserRepository
	ShelfRepository   *mocks.MockShelfRepository
	SectionRepository *mocks.MockSectionRepository
	LinkRepository    *mocks.MockLinkRepository
	SettingRepository *mocks.MockSettingRepository
}

func NewMockService(t *testing.T) *MockService {
	t.Helper()

	ctrl := gomock.NewController(t)

	userRepository := mocks.NewMockUserRepository(ctrl)
	shelfRepository := mocks.NewMockShelfRepository(ctrl)
	sectionRepository := mocks.NewMockSectionRepository(ctrl)
	linkRepository := mocks.NewMockLinkRepository(ctrl)
	settingService := mocks.NewMockSettingRepository(ctrl)

	repo := &repository.Repository{
		UserRepository:    userRepository,
		ShelfRepository:   shelfRepository,
		SectionRepository: sectionRepository,
		LinkRepository:    linkRepository,
		SettingRepository: settingService,
	}

	service := NewService(repo)

	return &MockService{
		Ctrl:              ctrl,
		Service:           service,
		UserRepository:    userRepository,
		ShelfRepository:   shelfRepository,
		SectionRepository: sectionRepository,
		LinkRepository:    linkRepository,
	}
}
