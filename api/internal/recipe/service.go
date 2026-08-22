package recipe

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	HasActiveReferences(ctx context.Context, difficultyID, durationID string) (bool, error)
	Create(ctx context.Context, recipe Recipe) (*Recipe, error)
	List(ctx context.Context, query GetRecipesQuery) ([]Recipe, int64, error)
	DifficultyExists(ctx context.Context, id string) (bool, error)
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

func (svc *service) List(ctx context.Context, query GetRecipesQuery) ([]Recipe, int64, error) {
	// Inject default parameter if blank
	query.Ensure()

	if query.Difficulty != "" {
		exists, err := svc.repository.DifficultyExists(ctx, query.Difficulty)
		if err != nil {
			return nil, 0, fmt.Errorf("list recipes: %w", err)

		}

		if !exists {
			return nil, 0, fmt.Errorf("%w: difficulty %q does not exist", ErrInvalidReferenceData, query.Difficulty)

		}
	}

	recipes, total, err := svc.repository.List(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("list recipes: %w", err)
	}

	return recipes, total, nil
}
