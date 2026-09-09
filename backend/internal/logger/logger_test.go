package logger

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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

func Test_Unit_Logger_Init_WritesJSONAtConfiguredLevel(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	original := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = original })

	Init("warn")
	zap.L().Info("this should be filtered out by the warn threshold")
	zap.L().Warn("this should appear", zap.String("k", "v"))

	require.NoError(t, w.Close())
	os.Stdout = original

	out, err := io.ReadAll(r)
	require.NoError(t, err)
	body := string(out)

	require.NotContains(t, body, "this should be filtered out")
	require.Contains(t, body, "this should appear")
	require.Contains(t, body, `"level":"WARN"`)
	require.Contains(t, body, `"k":"v"`)
}
