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

	code := m.Run()
	os.Exit(code)
}
