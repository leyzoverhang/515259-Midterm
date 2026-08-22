package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	oauthPrefix = "oauth_state:"
)

type repository struct {
	cache *redis.Client
}

func NewRepository(cache *redis.Client) *repository {
	return &repository{
		cache: cache,
	}
}

func (repo *repository) SaveState(ctx context.Context, state string, ttl time.Duration) error {
	return repo.cache.Set(ctx, (oauthPrefix + state), "1", ttl).Err()
}
