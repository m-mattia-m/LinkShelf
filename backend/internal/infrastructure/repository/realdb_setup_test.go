//go:build realdb

package repository

import (
	"backend/internal/config"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// TestRepository is a real Repository backed by a throwaway Postgres or
// MySQL container - whichever database.engine is configured (env
// APP_DATABASE_ENGINE) - migrated exactly the way NewRepository() migrates
// it in production. CI runs this test binary twice, once per engine (see
// .github/workflows/ci.yaml), so both engines' dialect-specific SQL is
// actually exercised, not just assumed compatible.
var TestRepository *Repository

func setupRealDB() (func(), error) {
	ctx := context.Background()
	engine := strings.ToLower(config.String("database.engine"))

	var terminate func(context.Context) error

	switch engine {
	case "mysql":
		ctr, err := mysql.Run(ctx, "mysql:8",
			mysql.WithDatabase(config.String("database.name")),
			mysql.WithUsername(config.String("database.username")),
			mysql.WithPassword(config.String("database.password")),
		)
		if err != nil {
			return nil, fmt.Errorf("start mysql container: %w", err)
		}
		terminate = func(ctx context.Context) error { return ctr.Terminate(ctx) }

		port, err := ctr.MappedPort(ctx, "3306/tcp")
		if err != nil {
			return nil, err
		}
		config.Set("database.host", "127.0.0.1")
		config.Set("database.port", port.Port())
		config.Set("database.params", "charset=utf8mb4&parseTime=true")

		dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/%s",
			config.String("database.username"),
			config.String("database.password"),
			port.Port(),
			config.String("database.name"),
		)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}
		defer db.Close()
		if err := waitForRealDatabase(ctx, db, 60*time.Second); err != nil {
			return nil, err
		}

	case "postgres":
		ctr, err := postgres.Run(ctx, "postgres:18",
			postgres.WithDatabase(config.String("database.name")),
			postgres.WithUsername(config.String("database.username")),
			postgres.WithPassword(config.String("database.password")),
		)
		if err != nil {
			return nil, fmt.Errorf("start postgres container: %w", err)
		}
		terminate = func(ctx context.Context) error { return ctr.Terminate(ctx) }

		port, err := ctr.MappedPort(ctx, "5432/tcp")
		if err != nil {
			return nil, err
		}
		config.Set("database.host", "127.0.0.1")
		config.Set("database.port", port.Port())

		dsn, err := ctr.ConnectionString(ctx)
		if err != nil {
			return nil, err
		}
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, err
		}
		defer db.Close()
		if err := waitForRealDatabase(ctx, db, 60*time.Second); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("realdb test tier does not support database.engine=%q", engine)
	}

	repo, err := NewRepository()
	if err != nil {
		return nil, fmt.Errorf("build repository against real %s: %w", engine, err)
	}
	TestRepository = repo

	return func() { _ = terminate(context.Background()) }, nil
}

func waitForRealDatabase(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("database not ready after %s", timeout)
		case <-ticker.C:
		}
	}
}
