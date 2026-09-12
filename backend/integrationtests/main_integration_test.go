//go:build integration
// +build integration

package integrationtests

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/controller"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

var (
	TestRepository *repository.Repository
	TestService    *domain.Service
	BaseURL        string
	httpServer     *http.Server
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Load config
	if err := config.LoadConfig(); err != nil {
		zap.L().Error(err.Error())
		panic(err)
	}

	// Start Postgres
	pg, err := postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase(config.String("database.name")),
		postgres.WithUsername(config.String("database.username")),
		postgres.WithPassword(config.String("database.password")),
	)
	if err != nil {
		panic(err)
	}

	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(err)
	}

	config.Set("database.host", "localhost")
	config.Set("database.port", port.Port())

	// Wait for DB
	dsn, _ := pg.ConnectionString(ctx)
	db, _ := sql.Open("pgx", dsn)
	if err := waitForDatabase(ctx, db, 30*time.Second); err != nil {
		panic(err)
	}

	// Init repository
	TestRepository, err = repository.NewRepository()
	if err != nil {
		panic(err)
	}

	TestService = domain.NewService(TestRepository, nil, nil)

	// Build router
	router, err := controller.Router(TestService)
	if err != nil {
		panic(err)
	}

	// Start HTTP server on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	BaseURL = "http://" + listener.Addr().String()

	httpServer = &http.Server{
		Handler: router,
	}

	go func() {
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Give server a moment
	time.Sleep(200 * time.Millisecond)

	// Run tests
	code := m.Run()

	// Teardown
	_ = httpServer.Shutdown(ctx)
	_ = pg.Terminate(ctx)

	os.Exit(code)
}

func waitForDatabase(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		err := db.PingContext(ctx)
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("database not ready after %s: %w", timeout, err)
		case <-ticker.C:
			// retry
		}
	}
}

// doRequest performs an unauthenticated request - fine for public endpoints,
// or for exercising a protected one's 401 path. Use doAuthedRequest for
// anything that needs a real caller identity.
func doRequest(
	t *testing.T,
	method, path string,
	body io.Reader,
) *http.Response {
	t.Helper()
	return doRequestWithToken(t, method, path, body, "")
}

// doAuthedRequest performs a request with the given access token attached as
// a Bearer Authorization header.
func doAuthedRequest(
	t *testing.T,
	method, path string,
	body io.Reader,
	token string,
) *http.Response {
	t.Helper()
	return doRequestWithToken(t, method, path, body, token)
}

func doRequestWithToken(
	t *testing.T,
	method, path string,
	body io.Reader,
	token string,
) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, BaseURL+path, body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return resp
}

func ObjectToJSON(object any) string {
	bytes, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

// loginAndGetToken exercises the real login endpoint and returns the access
// token, so tests act as a genuine authenticated caller rather than reaching
// around auth entirely.
func loginAndGetToken(t *testing.T, email, password string) string {
	t.Helper()

	resp := doRequest(t, http.MethodPost, "/v1/auth/login", strings.NewReader(ObjectToJSON(model.LoginRequest{
		Email:    email,
		Password: password,
	})))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var tokens model.TokenPair
	require.NoError(t, json.Unmarshal(body, &tokens))
	require.NotEmpty(t, tokens.AccessToken)

	return tokens.AccessToken
}

// createUserWithRole creates a user directly via the service layer (bypassing
// HTTP, since self-registration can never grant a role), then logs in over
// the real HTTP API so the test gets back a genuine bearer token to act as
// that user.
func createUserWithRole(t *testing.T, role string) (userId, token string) {
	t.Helper()

	randUuid, err := uuid.NewV7()
	require.NoError(t, err)

	email := fmt.Sprintf("test-user-%s@test.com", ShortUUID(randUuid.String()))
	const password = "secret"

	user, err := TestService.UserService.Create(&model.UserCreate{
		UserBase: model.UserBase{
			Email:     email,
			FirstName: "test-firstname",
			LastName:  "test-lastname",
			Role:      role,
		},
		Password: password,
	}, true)
	require.NoError(t, err)

	return user.Id, loginAndGetToken(t, email, password)
}

func createTestUser(t *testing.T) (userId, token string) {
	t.Helper()
	return createUserWithRole(t, model.RoleUser)
}

func createTestAdmin(t *testing.T) (userId, token string) {
	t.Helper()
	return createUserWithRole(t, model.RoleAdmin)
}

func getShelfInclusiveItsOwnerUser(t *testing.T) (shelfId, token string) {
	t.Helper()

	userId, token := createTestUser(t)

	randUuid, err := uuid.NewV7()
	require.NoError(t, err)

	shelfId, err = TestService.ShelfService.Create(userId, &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       fmt.Sprintf("shelf-for-owner-%s", ShortUUID(randUuid.String())),
			Path:        fmt.Sprintf("shelf-for-owner-%s", ShortUUID(randUuid.String())),
			Description: "A shelf created during API integration tests",
			Icon:        "",
		},
		Theme: "",
	})
	require.NoError(t, err)

	return shelfId, token
}

func getSectionAndShelfInclusiveItsOwnerUser(t *testing.T) (sectionId, token string) {
	t.Helper()

	shelfId, token := getShelfInclusiveItsOwnerUser(t)

	section, err := TestService.SectionService.Create("", true, &model.Section{
		SectionBase: model.SectionBase{
			Title:   "test-section-get",
			ShelfId: shelfId,
		},
	})
	require.NoError(t, err)

	return section.Id, token
}

func ShortUUID(u string) string {
	parts := strings.Split(u, "-")
	return parts[len(parts)-1]
}
