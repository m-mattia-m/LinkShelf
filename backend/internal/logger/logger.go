package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init builds a JSON-structured Zap logger at the given level ("debug", "info",
// "warn", "error" or "fatal") and installs it as the global logger, accessible
// anywhere via zap.L() / zap.S().
func Init(levelStr string) {
	level := ParseLevel(levelStr)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)

	zap.L().Info("Logger initialized", zap.String("level", level.String()))
}

// ParseLevel maps a config string ("debug", "info", "warn", "error", "fatal")
// to a zap level, defaulting to info for anything unrecognized.
func ParseLevel(levelStr string) zapcore.Level {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		return zapcore.InfoLevel
	}
	return level
}
