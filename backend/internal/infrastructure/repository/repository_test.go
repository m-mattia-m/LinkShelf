package repository

import (
	"backend/internal/config"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	config.Reset()

	// Ensure test config is loaded
	if err := config.LoadConfig(); err != nil {
		panic("failed to load test config: " + err.Error())
	}

	// setupRealDB is a no-op unless built with -tags=realdb, in which case it
	// spins up a real Postgres or MySQL container (per database.engine) and
	// points TestRepository at it - see realdb_setup_test.go.
	cleanup, err := setupRealDB()
	if err != nil {
		panic("failed to set up realdb test tier: " + err.Error())
	}

	code := m.Run()
	cleanup()
	os.Exit(code)
}
