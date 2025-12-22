package config

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Unit_Config_Logging_ParseLevel_Debug(t *testing.T) {
	require.Equal(t, ParseLevel("debug"), slog.LevelDebug)
}

func Test_Unit_Config_Logging_ParseLevel_Info(t *testing.T) {
	require.Equal(t, ParseLevel("info"), slog.LevelInfo)
}

func Test_Unit_Config_Logging_ParseLevel_Warn(t *testing.T) {
	require.Equal(t, ParseLevel("warn"), slog.LevelWarn)
}

func Test_Unit_Config_Logging_ParseLevel_Warning(t *testing.T) {
	require.Equal(t, ParseLevel("warning"), slog.LevelWarn)
}

func Test_Unit_Config_Logging_ParseLevel_Error(t *testing.T) {
	require.Equal(t, ParseLevel("error"), slog.LevelError)
}

func Test_Unit_Config_Logging_ParseLevel_Default(t *testing.T) {
	require.Equal(t, ParseLevel("unknown"), slog.LevelInfo)
}
