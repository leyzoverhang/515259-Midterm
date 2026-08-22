package config

import (
	"fmt"
	"time"
)

type App struct {
	Name            string        `env:"APP_NAME" envDefault:"wongnok"`
	Env             string        `env:"APP_ENV" envDefault:"development"`
	Port            int           `env:"APP_PORT" envDefault:"8080"`
	ShutdownTimeout time.Duration `env:"APP_SHUTDOWN" envDefault:"10s"`
}

func (app App) Addr() string {
	return fmt.Sprintf(":%d", app.Port)
}

func (app App) IsProduction() bool {
	return app.Env == "production"
}
