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

	// setupRealDB starts a real database only with -tags=realdb.
	cleanup, err := setupRealDB()
	if err != nil {
		panic("failed to set up realdb test tier: " + err.Error())
	}

	code := m.Run()
	cleanup()
	os.Exit(code)
}
