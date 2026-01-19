package config

import (
	"log"
	"log/slog"
	"os"
	"strings"
)

const (
	LevelDebug = slog.Level(-4)
	LevelInfo  = slog.Level(0)
	LevelWarn  = slog.Level(4)
	LevelError = slog.Level(8)
)

func InitLogging(lvl string) {
	level := toLevel(lvl)
	log.SetOutput(os.Stdout)
	slog.SetLogLoggerLevel(level)
}

func toLevel(lvl string) slog.Level {
	levels := map[string]slog.Level{
		"debug": LevelDebug,
		"info":  LevelInfo,
		"warn":  LevelWarn,
		"error": LevelError,
	}
	if level, ok := levels[strings.ToLower(lvl)]; ok {
		return level
	}
	return LevelInfo
}
