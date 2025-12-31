package repository

import (
	"backend/migrations"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/spf13/viper"
)

type Repository struct {
	UserRepository    UserRepository
	ShelfRepository   ShelfRepository
	SectionRepository SectionRepository
	LinkRepository    LinkRepository
	SettingRepository SettingRepository
}

func NewRepository() (*Repository, error) {
	sqlDSN, driver, migrateDSN, err := getConnectionInformation()
	if err != nil {
		return nil, err
	}

	db, err := connectToDatabase(sqlDSN, driver)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(migrateDSN); err != nil {
		return nil, err
	}

	userRepo, err := NewUserRepository(db, "user")
	if err != nil {
		return nil, err
	}

	shelfRepo, err := NewShelfRepository(db, "shelf")
	if err != nil {
		return nil, err
	}

	sectionRepo, err := NewSectionRepository(db, "section")
	if err != nil {
		return nil, err
	}

	linkRepo, err := NewLinkRepository(db, "link")
	if err != nil {
		return nil, err
	}

	settingRepo, err := NewSettingRepository(db, "setting")
	if err != nil {
		return nil, err
	}

	return &Repository{
		UserRepository:    userRepo,
		ShelfRepository:   shelfRepo,
		SectionRepository: sectionRepo,
		LinkRepository:    linkRepo,
		SettingRepository: settingRepo,
	}, nil
}

func connectToDatabase(dsn, driver string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Validate connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	slog.Info("Database connected", slog.String("driver", driver))
	return db, nil
}

func runMigrations(migrateDSN string) error {
	slog.Info("Applying DB migrations...")

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migration source failed: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		source,
		migrateDSN,
	)
	if err != nil {
		return fmt.Errorf("migration setup failed: %w", err)
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		slog.Info("No new migrations")
		return nil
	}
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	slog.Info("Migrations applied successfully")
	return nil
}

func getConnectionInformation() (sqlDSN, driver, migrateDSN string, err error) {
	engine := strings.ToLower(viper.GetString("database.engine"))

	host := viper.GetString("database.host")
	port := viper.GetString("database.port")
	dbname := viper.GetString("database.name")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")
	params := viper.GetString("database.params")

	safePassword := "***"

	switch engine {
	case "postgres":
		driver = "pgx"

		// database/sql DSN (NO scheme)
		sqlDSN = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s",
			host, port, username, password, dbname,
		)

		// Migrate DB URL DSN
		migrateDSN = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?%s",
			username, password, host, port, dbname, params,
		)
	case "mysql":
		driver = "mysql"

		// database/sql DSN (NO scheme)
		sqlDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			username, password, host, port, dbname)

		migrateDSN = fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s?%s",
			username, password, host, port, dbname, params)
	default:
		slog.Error("Unsupported DB engine", slog.String("engine", engine))
		os.Exit(1)
	}

	sqlDSN = strings.TrimSuffix(sqlDSN, "?")
	migrateDSN = strings.TrimSuffix(migrateDSN, "?")

	// Debug log with masked password
	slog.Debug("SQL DSN: " + strings.Replace(sqlDSN, password, safePassword, -1))
	slog.Debug("Migration DSN: " + strings.Replace(migrateDSN, password, safePassword, -1))

	return sqlDSN, driver, migrateDSN, nil
}

func buildSqlStatements(query string) (string, error) {
	_, driver, _, err := getConnectionInformation()
	if err != nil {
		return "", err
	}

	if driver != "pgx" {
		return query, nil
	}

	next := 1
	res := make([]rune, 0, len(query))

	for _, r := range query {
		if r == '?' {
			res = append(res, []rune(fmt.Sprintf("$%d", next))...)
			next++
		} else {
			res = append(res, r)
		}
	}
	return string(res), nil
}
