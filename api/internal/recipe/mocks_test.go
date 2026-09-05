package recipe

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a hand-written stand-in for what `mockery` would
// generate from the Repository interface (see .mockery.yml). ต้องรัน
// `mockery` จริงทุกครั้งที่แก้ interface — ไฟล์นี้เขียนมือไว้ก่อนเพื่อให้เทสรันได้
// ทันที ถ้ามี mockery ในเครื่องแล้ว ลบไฟล์นี้แล้วรัน `mockery` แทนได้เลย
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) HasActiveReferences(ctx context.Context, difficultyID, durationID string) (bool, error) {
	args := m.Called(ctx, difficultyID, durationID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) Create(ctx context.Context, recipe Recipe) (*Recipe, error) {
	args := m.Called(ctx, recipe)

	var result *Recipe
	if v := args.Get(0); v != nil {
		result = v.(*Recipe)
	}

	return result, args.Error(1)
}

func (m *MockRepository) FindByID(ctx context.Context, id int) (*Recipe, error) {
	args := m.Called(ctx, id)

	var result *Recipe
	if v := args.Get(0); v != nil {
		result = v.(*Recipe)
	}

	return result, args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, userID uuid.UUID, query GetRecipesQuery) ([]Recipe, int64, error) {
	args := m.Called(ctx, userID, query)

	var result []Recipe
	if v := args.Get(0); v != nil {
		result = v.([]Recipe)
	}

	return result, args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) DifficultyExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) Favorite(ctx context.Context, userID uuid.UUID, recipeID int) error {
	args := m.Called(ctx, userID, recipeID)
	return args.Error(0)
}

func (m *MockRepository) Unfavorite(ctx context.Context, userID uuid.UUID, recipeID int) error {
	args := m.Called(ctx, userID, recipeID)
	return args.Error(0)
}

func (m *MockRepository) Rate(ctx context.Context, userID uuid.UUID, recipeID int, score float64) error {
	args := m.Called(ctx, userID, recipeID, score)
	return args.Error(0)
}

func (m *MockRepository) FavoriteRecipeIDs(ctx context.Context, userID uuid.UUID, recipeIDs []int) (map[int]bool, error) {
	args := m.Called(ctx, userID, recipeIDs)

	var result map[int]bool
	if v := args.Get(0); v != nil {
		result = v.(map[int]bool)
	}

	return result, args.Error(1)
}

func (m *MockRepository) RatingCounts(ctx context.Context, recipeIDs []int) (map[int]int64, error) {
	args := m.Called(ctx, recipeIDs)

	var result map[int]int64
	if v := args.Get(0); v != nil {
		result = v.(map[int]int64)
	}

	return result, args.Error(1)
}

// MockService is a hand-written stand-in for the handler package's Service interface.
type MockService struct {
	mock.Mock
}

func (m *MockService) Create(ctx context.Context, creatorID uuid.UUID, recipe Recipe) (*Recipe, error) {
	args := m.Called(ctx, creatorID, recipe)

	var result *Recipe
	if v := args.Get(0); v != nil {
		result = v.(*Recipe)
	}

	return result, args.Error(1)
}

func (m *MockService) FindByID(ctx context.Context, id int) (*Recipe, error) {
	args := m.Called(ctx, id)

	var result *Recipe
	if v := args.Get(0); v != nil {
		result = v.(*Recipe)
	}

	return result, args.Error(1)
}

func (m *MockService) List(ctx context.Context, userID uuid.UUID, query GetRecipesQuery) ([]RecipeView, int64, error) {
	args := m.Called(ctx, userID, query)

	var result []RecipeView
	if v := args.Get(0); v != nil {
		result = v.([]RecipeView)
	}

	return result, args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetByID(ctx context.Context, userID uuid.UUID, id int) (*RecipeView, error) {
	args := m.Called(ctx, userID, id)

	var result *RecipeView
	if v := args.Get(0); v != nil {
		result = v.(*RecipeView)
	}

	return result, args.Error(1)
}

func (m *MockService) Favorite(ctx context.Context, userID uuid.UUID, recipeID int) error {
	args := m.Called(ctx, userID, recipeID)
	return args.Error(0)
}

func (m *MockService) Unfavorite(ctx context.Context, userID uuid.UUID, recipeID int) error {
	args := m.Called(ctx, userID, recipeID)
	return args.Error(0)
}

func (m *MockService) Rate(ctx context.Context, userID uuid.UUID, recipeID int, score float64) error {
	args := m.Called(ctx, userID, recipeID, score)
	return args.Error(0)
}
