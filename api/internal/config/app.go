package config

import (
	"fmt"
	"time"
)

type App struct {
	Name            string        `env:"APP_NAME" envDefault:"wongnok" validate:"required"`
	Env             string        `env:"APP_ENV" envDefault:"development" validate:"required,oneof=development staging production"`
	Port            int           `env:"APP_PORT" envDefault:"8080" validate:"gte=1,lte=65535"`
	ShutdownTimeout time.Duration `env:"APP_SHUTDOWN" envDefault:"10s" validate:"required"`
}

func (app App) Addr() string {
	return fmt.Sprintf(":%d", app.Port)
}

func (app App) IsProduction() bool {
	return app.Env == "production"
}

func (app App) Validate() error {
	return validate.Struct(app)
}
