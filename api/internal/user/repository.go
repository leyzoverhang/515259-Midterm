package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type repository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewRepository(db *gorm.DB, cache *redis.Client) *repository {
	return &repository{
		db:    db,
		cache: cache,
	}
}

func (repo *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var result User

	if err := repo.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &result, nil
}

func (repo *repository) FindByUID(ctx context.Context, uid string) (*User, error) {
	var result User

	if err := repo.db.WithContext(ctx).First(&result, "uid = ?", uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &result, nil
}

func (repo *repository) Create(ctx context.Context, user User) error {
	return repo.db.WithContext(ctx).Create(&user).Error
}

func (repo *repository) Update(ctx context.Context, user User) error {
	return repo.db.WithContext(ctx).Save(&user).Error
}

func (repo *repository) ResolveID(ctx context.Context, uid string) (uuid.UUID, error) {
	cacheKey := "user_resolve_id:" + uid

	if cached, err := repo.cache.Get(ctx, cacheKey).Result(); err != nil {
		if id, err := uuid.Parse(cached); err == nil {
			return id, nil
		}
	}

	// Create cache
	user, err := repo.FindByUID(ctx, uid)
	if err != nil {
		return uuid.Nil, fmt.Errorf("find user by uid: %w", err)
	}

	repo.cache.Set(ctx, cacheKey, user.ID.String(), (24 * time.Hour))

	return user.ID, nil
}
