package user

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wongnok/internal/convutil"

	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByUID(ctx context.Context, uid string) (*User, error)
	Create(ctx context.Context, user User) error
	Update(ctx context.Context, user User) error
	ResolveID(ctx context.Context, uid string) (uuid.UUID, error)
}

type service struct {
	repository Repository
}

func NewService(repo Repository) *service {
	return &service{
		repository: repo,
	}
}

func (svc *service) FindByID(ctx context.Context, uid uuid.UUID) (*User, error) {
	return svc.repository.FindByID(ctx, uid)
}

func (svc *service) UpsertFromKeycloak(ctx context.Context, kuser KeycloakUser) error {
	// Find user by uid
	existing, err := svc.repository.FindByUID(ctx, kuser.UID)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return fmt.Errorf("find user: %w", err)
	}

	now := time.Now()

	// If not exists then create user
	if errors.Is(err, ErrUserNotFound) {
		return svc.repository.Create(ctx, User{
			ID:                uuid.New(),
			UID:               kuser.UID,
			Email:             kuser.Email,
			Name:              convutil.ToPointer(kuser.Name),
			PreferredUsername: convutil.ToPointer(kuser.PreferredUsername),
			LastSignedInAt:    convutil.ToPointer(now),
		})
	}

	// If exists update user
	existing.Email = kuser.Email
	existing.Name = convutil.ToPointer(kuser.Name)
	existing.PreferredUsername = convutil.ToPointer(kuser.PreferredUsername)
	existing.LastSignedInAt = convutil.ToPointer(now)

	return svc.repository.Update(ctx, *existing)
}

func (svc *service) ResolveID(ctx context.Context, uid string) (uuid.UUID, error) {
	return svc.repository.ResolveID(ctx, uid)
}
