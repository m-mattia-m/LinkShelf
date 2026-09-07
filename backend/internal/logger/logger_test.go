package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func Test_Unit_Logger_ParseLevel_Debug(t *testing.T) {
	require.Equal(t, zapcore.DebugLevel, ParseLevel("debug"))
}

func Test_Unit_Logger_ParseLevel_Info(t *testing.T) {
	require.Equal(t, zapcore.InfoLevel, ParseLevel("info"))
}

func Test_Unit_Logger_ParseLevel_Warn(t *testing.T) {
	require.Equal(t, zapcore.WarnLevel, ParseLevel("warn"))
}

func Test_Unit_Logger_ParseLevel_Error(t *testing.T) {
	require.Equal(t, zapcore.ErrorLevel, ParseLevel("error"))
}

func Test_Unit_Logger_ParseLevel_Default(t *testing.T) {
	require.Equal(t, zapcore.InfoLevel, ParseLevel("unknown"))
}
