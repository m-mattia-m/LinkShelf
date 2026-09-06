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

	TestService = domain.NewService(TestRepository, nil)

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

func doRequest(
	t *testing.T,
	method, path string,
	body io.Reader,
) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, BaseURL+path, body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

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

func getShelfOwnerUser() (string, error) {

	randUuid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	userRequest := &model.UserCreate{
		UserBase: model.UserBase{
			Email:     fmt.Sprintf("test-shelf-owner-%s@test.com", ShortUUID(randUuid.String())),
			FirstName: "test-shelf-owner-firstname",
			LastName:  "test-shelf-owner-lastname",
		},
		Password: "secret",
	}

	user, err := TestService.UserService.Create(userRequest)
	if err != nil {
		return "", err
	}

	return user.Id, nil
}

func getShelfInclusiveItsOwnerUser() (string, error) {
	userId, err := getShelfOwnerUser()
	if err != nil {
		return "", err
	}

	randUuid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	shelfId, err := TestService.ShelfService.Create(userId, &model.Shelf{
		PublicShelf: model.PublicShelf{
			Title:       fmt.Sprintf("shelf-for-owner-%s", ShortUUID(randUuid.String())),
			Path:        fmt.Sprintf("shelf-for-owner-%s", ShortUUID(randUuid.String())),
			Description: "A shelf created during API integration tests",
			Icon:        "",
		},
		Theme: "",
	})
	if err != nil {
		return "", err
	}

	return shelfId, nil

}

func getSectionAndShelfInclusiveItsOwnerUser() (string, error) {
	shelfId, err := getShelfInclusiveItsOwnerUser()
	if err != nil {
		return "", err
	}

	section, err := TestService.SectionService.Create("", true, &model.Section{
		SectionBase: model.SectionBase{
			Title:   "test-section-get",
			ShelfId: shelfId,
		},
	})

	return section.Id, nil
}

func ShortUUID(u string) string {
	parts := strings.Split(u, "-")
	return parts[len(parts)-1]
}
