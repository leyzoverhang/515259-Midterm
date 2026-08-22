package config

import (
	"errors"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Config struct {
	App      App
	Database Database
	Logging  Logging
	Keycloak Keycloak
	Redis    Redis
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	return errors.Join(
		cfg.App.Validate(),
		cfg.Database.Validate(),
		cfg.Logging.Validate(),
		cfg.Keycloak.Validate(),
		cfg.Redis.Validate(),
	)
}
