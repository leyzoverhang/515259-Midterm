package config

type Database struct {
	PostgresDSN string `env:"POSTGRES_DSN"`
}
