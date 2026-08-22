package config

import (
	"log/slog"
	"strings"
)

type Logging struct {
	Level  string `env:"LOG_LEVEL" envDefault:"DEBUG"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

func (Logging Logging) SlogLevel() slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(strings.TrimSpace(Logging.Level)))); err != nil {
		return slog.LevelInfo
	}

	return level
}
