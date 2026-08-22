package auth

import "github.com/redis/go-redis/v9"

type repository struct {
	cache *redis.Client
}

func NewRepository(cache *redis.Client) *repository {
	return &repository{
		cache: cache,
	}
}
