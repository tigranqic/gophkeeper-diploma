// Package logger provides initialization and access to a zap.Logger.
package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init creates and stores a global zap.Logger configured with the given level
// ("debug", "info", "warn", "error") and format ("json" for production, any
// other value for human-readable development output).
func Init(levelStr, formatStr string) *zap.Logger {
	level := parseLevel(levelStr)
	format := strings.ToLower(strings.TrimSpace(formatStr))

	var cfg zap.Config

	switch format {
	case "json":
		cfg = zap.NewProductionConfig()
	default:
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.Level = zap.NewAtomicLevelAt(level)

	var err error
	log, err = cfg.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		log = zap.NewNop()
		return log
	}

	log.Info("logger initialized",
		zap.String("level", level.String()),
		zap.String("format", formatOrDefault(format)))

	return log
}

func parseLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func formatOrDefault(format string) string {
	if format == "" {
		return "text"
	}
	return format
}

// Get returns the global logger initialised by Init. If Init has not been
// called yet, a no-op logger is returned so callers never receive a nil pointer.
func Get() *zap.Logger {
	if log == nil {
		return zap.NewNop()
	}
	return log
}
