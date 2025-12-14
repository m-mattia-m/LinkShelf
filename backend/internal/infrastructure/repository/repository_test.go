package repository

import (
	"backend/internal/config"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testRepo *Repository
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	if err := config.LoadConfig(); err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	pg, err := postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase(viper.GetString("database.name")),
		postgres.WithUsername(viper.GetString("database.username")),
		postgres.WithPassword(viper.GetString("database.password")),
	)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	viper.Set("database.host", "localhost")
	viper.Set("database.port", port.Port())

	dsn, err := pg.ConnectionString(ctx)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	err = waitForDatabase(ctx, db, 30*time.Second)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	testRepo, err = NewRepository()
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	code := m.Run()

	err = pg.Terminate(ctx)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

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
