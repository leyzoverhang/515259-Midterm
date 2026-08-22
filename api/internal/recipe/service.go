package recipe

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	HasActiveReferences(ctx context.Context, difficultyID, durationID string) (bool, error)
	Create(ctx context.Context, recipe Recipe) (*Recipe, error)
}

type service struct {
	repository Repository
}

func NewService(repo Repository) *service {
	return &service{
		repository: repo,
	}
}

func (svc *service) Create(ctx context.Context, creatorID uuid.UUID, recipe Recipe) (*Recipe, error) {
	refActive, err := svc.repository.HasActiveReferences(ctx, recipe.DifficultyID, recipe.DurationID)
	if err != nil {
		return nil, err
	}

	if !refActive {
		return nil, ErrInvalidReferenceData
	}

	recipe.CreatorID = creatorID

	return svc.repository.Create(ctx, recipe)
}
