package main

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/controller"
	"backend/internal/infrastructure/repository"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	TestRepository *repository.Repository
	BaseURL        string
	httpServer     *http.Server
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Load config
	if err := config.LoadConfig(); err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	// Start Postgres
	pg, err := postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase(viper.GetString("database.name")),
		postgres.WithUsername(viper.GetString("database.username")),
		postgres.WithPassword(viper.GetString("database.password")),
	)
	if err != nil {
		panic(err)
	}

	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(err)
	}

	viper.Set("database.host", "localhost")
	viper.Set("database.port", port.Port())

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

	TestService := domain.NewService(TestRepository)

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

	resp, err := http.DefaultClient.Do(req)
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
