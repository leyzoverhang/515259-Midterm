package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, user User) (*User, error)
}

type service struct {
	repository Repository
}

func NewService(repo Repository) *service {
	return &service{
		repository: repo,
	}
}

func (svc *service) FindByID(ctx context.Context, id string) (*User, error) {
	parsedID, err := uuid.Parse((id))
	if err != nil {
		return nil, ErrInvalidInput
	}

	return svc.repository.FindByID(ctx, parsedID)
}

func (svc *service) Create(ctx context.Context, user User) (*User, error) {
	return svc.repository.Create(ctx, user)
}
