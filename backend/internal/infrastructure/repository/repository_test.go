package repository

import (
	"backend/internal/config"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	// Reset viper global state (important!)
	viper.Reset()

	// Ensure test config is loaded
	if err := config.LoadConfig(); err != nil {
		panic("failed to load test config: " + err.Error())
	}

	code := m.Run()
	os.Exit(code)
}
