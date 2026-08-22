package config

import "fmt"

type Redis struct {
	Host     string `env:"REDIS_HOST" envDefault:"localhost" validate:"required"`
	Port     int    `env:"REDIS_PORT" envDefault:"6380" validate:"gte=1,lte=65535"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0" validate:"gte=0,lte=15"`
}

func (redis Redis) Addr() string {
	return fmt.Sprintf("%s:%d", redis.Host, redis.Port)
}

func (redis Redis) Validate() error {
	return validate.Struct(redis)
}
