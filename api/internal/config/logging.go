package config

import (
	"log/slog"
	"strings"
)

type Logging struct {
	Level  string `env:"LOG_LEVEL" envDefault:"DEBUG" validate:"required,oneofci=DEBUG INFO WARN ERROR"`
	Format string `env:"LOG_FORMAT" envDefault:"json" validate:"required,oneof=json text"`
}

func (logging Logging) SlogLevel() slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(strings.TrimSpace(logging.Level)))); err != nil {
		return slog.LevelInfo
	}

	return level
}

func (logging Logging) Validate() error {
	return validate.Struct(logging)
}
