package controller

import (
	"backend/internal/domain"
	"backend/internal/domain/mocks"
	"testing"

	"go.uber.org/mock/gomock"
)

type MockService struct {
	Ctrl    *gomock.Controller
	Service *domain.Service

	UserService      *mocks.MockUserService
	ShelfService     *mocks.MockShelfService
	SectionService   *mocks.MockSectionService
	LinkService      *mocks.MockLinkService
	StatisticService *mocks.MockStatisticService
}

func NewMockDomainService(t *testing.T) *MockService {
	t.Helper()

	ctrl := gomock.NewController(t)

	userService := mocks.NewMockUserService(ctrl)
	shelfService := mocks.NewMockShelfService(ctrl)
	sectionService := mocks.NewMockSectionService(ctrl)
	linkService := mocks.NewMockLinkService(ctrl)
	statisticService := mocks.NewMockStatisticService(ctrl)

	domainService := &domain.Service{
		UserService:      userService,
		ShelfService:     shelfService,
		SectionService:   sectionService,
		LinkService:      linkService,
		StatisticService: statisticService,
	}

	return &MockService{
		Ctrl:             ctrl,
		Service:          domainService,
		UserService:      userService,
		ShelfService:     shelfService,
		SectionService:   sectionService,
		LinkService:      linkService,
		StatisticService: statisticService,
	}
}
