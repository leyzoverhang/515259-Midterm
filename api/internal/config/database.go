package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Database struct {
	PostgresDSN string `env:"POSTGRES_DSN" validate:"required"`
}

func (db Database) Validate() error {
	if err := validate.Struct(db); err != nil {
		return err
	}

	uri, err := url.Parse(db.PostgresDSN)
	if err != nil {
		return fmt.Errorf("POSTGRES_DSN: %w", err)
	}

	if strings.Trim(uri.Path, "/") == "" && uri.Query().Get("database") == "" {
		return errors.New("POSTGRES_DSN: ไม่มีชื่อ database")
	}

	return nil
}
