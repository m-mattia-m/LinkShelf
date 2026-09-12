package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/domain/mocks"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type MockService struct {
	Ctrl    *gomock.Controller
	Service *domain.Service

	UserService              *mocks.MockUserService
	ShelfService             *mocks.MockShelfService
	SectionService           *mocks.MockSectionService
	LinkService              *mocks.MockLinkService
	StatisticService         *mocks.MockStatisticService
	SettingService           *mocks.MockSettingService
	AuthService              *mocks.MockAuthService
	EmailVerificationService *mocks.MockEmailVerificationService
}

func NewMockDomainService(t *testing.T) *MockService {
	t.Helper()

	ctrl := gomock.NewController(t)

	userService := mocks.NewMockUserService(ctrl)
	shelfService := mocks.NewMockShelfService(ctrl)
	sectionService := mocks.NewMockSectionService(ctrl)
	linkService := mocks.NewMockLinkService(ctrl)
	statisticService := mocks.NewMockStatisticService(ctrl)
	settingService := mocks.NewMockSettingService(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	emailVerificationService := mocks.NewMockEmailVerificationService(ctrl)

	domainService := &domain.Service{
		UserService:              userService,
		ShelfService:             shelfService,
		SectionService:           sectionService,
		LinkService:              linkService,
		StatisticService:         statisticService,
		SettingService:           settingService,
		AuthService:              authService,
		EmailVerificationService: emailVerificationService,
	}

	return &MockService{
		Ctrl:                     ctrl,
		Service:                  domainService,
		UserService:              userService,
		ShelfService:             shelfService,
		SectionService:           sectionService,
		LinkService:              linkService,
		StatisticService:         statisticService,
		SettingService:           settingService,
		AuthService:              authService,
		EmailVerificationService: emailVerificationService,
	}
}

func Test_Router_BuildsWithoutError(t *testing.T) {
	config.Reset()
	require.NoError(t, config.LoadConfig())

	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	router, err := Router(svc.Service)

	require.NoError(t, err)
	require.NotNil(t, router)
}
