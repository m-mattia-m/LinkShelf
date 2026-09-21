package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/domain/mocks"
	"backend/internal/infrastructure/api/model"
	"net/http"
	"net/http/httptest"
	"strings"
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
	ThemeService             *mocks.MockThemeService
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
	themeService := mocks.NewMockThemeService(ctrl)
	statisticService := mocks.NewMockStatisticService(ctrl)
	settingService := mocks.NewMockSettingService(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	emailVerificationService := mocks.NewMockEmailVerificationService(ctrl)

	domainService := &domain.Service{
		UserService:              userService,
		ShelfService:             shelfService,
		SectionService:           sectionService,
		LinkService:              linkService,
		ThemeService:             themeService,
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
		ThemeService:             themeService,
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

// Both public lookups are registered side by side, so this goes through the
// real router: the two paths must not collide in gin, and each must reach its
// own handler.
func Test_Router_ServesBothPublicShelfLookups(t *testing.T) {
	config.Reset()
	require.NoError(t, config.LoadConfig())

	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	svc.ShelfService.EXPECT().
		GetByPath("my-path").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "by-path", Path: "my-path"}}, nil)
	svc.ShelfService.EXPECT().
		GetByUsernameAndPath("alice", "my-path").
		Return(&model.Shelf{PublicShelf: model.PublicShelf{Id: "by-user", Path: "my-path"}}, nil)

	router, err := Router(svc.Service)
	require.NoError(t, err)

	for url, wantId := range map[string]string{
		"/v1/shelves/by-path/my-path":       "by-path",
		"/v1/shelves/by-user/alice/my-path": "by-user",
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))

		require.Equal(t, http.StatusOK, rec.Code, url)
		require.Contains(t, rec.Body.String(), `"id":"`+wantId+`"`, url)
	}
}

func Test_Router_PublishesTheUsernameInTheOpenApiSchema(t *testing.T) {
	config.Reset()
	require.NoError(t, config.LoadConfig())

	svc := NewMockDomainService(t)
	defer svc.Ctrl.Finish()

	router, err := Router(svc.Service)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	spec := rec.Body.String()
	require.Contains(t, spec, "/v1/shelves/by-user/{username}/{path}")
	require.Contains(t, spec, "get-public-shelf-by-username-and-path")
	require.Contains(t, spec, `"user_based_paths"`)
	// The schema documents the username shape, so clients can pre-validate.
	require.True(t, strings.Contains(spec, `"username"`) && strings.Contains(spec, `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`))
}
