package recipe

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// เทสชั้น service ด้วย mock — พิสูจน์ "กฎธุรกิจ" (ตามหมวด 07 ของคู่มือ)
// ไม่แตะ DB จริง เพราะ Repository ถูก mock ไว้หมดแล้ว

func TestService_Favorite_ExistingRecipe_CallsRepository(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	recipeID := 42

	repo.On("FindByID", ctx, recipeID).Return(&Recipe{ID: recipeID}, nil)
	repo.On("Favorite", ctx, userID, recipeID).Return(nil)

	err := svc.Favorite(ctx, userID, recipeID)

	require.NoError(t, err)
	repo.AssertCalled(t, "Favorite", ctx, userID, recipeID)
}

func TestService_Favorite_RecipeNotFound_DoesNotTouchRepository(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	recipeID := 999

	repo.On("FindByID", ctx, recipeID).Return(nil, ErrRecipeNotFound)

	err := svc.Favorite(ctx, userID, recipeID)

	assert.ErrorIs(t, err, ErrRecipeNotFound)
	repo.AssertNotCalled(t, "Favorite", mock.Anything, mock.Anything, mock.Anything)
}

func TestService_Favorite_RepositoryError_Propagates(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	recipeID := 42
	repoErr := errors.New("connection refused")

	repo.On("FindByID", ctx, recipeID).Return(&Recipe{ID: recipeID}, nil)
	repo.On("Favorite", ctx, userID, recipeID).Return(repoErr)

	err := svc.Favorite(ctx, userID, recipeID)

	assert.ErrorIs(t, err, repoErr)
}

func TestService_Unfavorite_ExistingRecipe_CallsRepository(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	recipeID := 42

	repo.On("FindByID", ctx, recipeID).Return(&Recipe{ID: recipeID}, nil)
	repo.On("Unfavorite", ctx, userID, recipeID).Return(nil)

	err := svc.Unfavorite(ctx, userID, recipeID)

	require.NoError(t, err)
	repo.AssertCalled(t, "Unfavorite", ctx, userID, recipeID)
}

func TestService_Unfavorite_RecipeNotFound_DoesNotTouchRepository(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	recipeID := 999

	repo.On("FindByID", ctx, recipeID).Return(nil, ErrRecipeNotFound)

	err := svc.Unfavorite(ctx, userID, recipeID)

	assert.ErrorIs(t, err, ErrRecipeNotFound)
	repo.AssertNotCalled(t, "Unfavorite", mock.Anything, mock.Anything, mock.Anything)
}

func TestService_Rate_ExistingRecipe_CallsRepositoryWithScore(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	recipeID := 7

	repo.On("FindByID", ctx, recipeID).Return(&Recipe{ID: recipeID}, nil)
	repo.On("Rate", ctx, userID, recipeID, 4.5).Return(nil)

	err := svc.Rate(ctx, userID, recipeID, 4.5)

	require.NoError(t, err)
	repo.AssertCalled(t, "Rate", ctx, userID, recipeID, 4.5)
}

func TestService_Rate_RecipeNotFound_DoesNotTouchRepository(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	recipeID := 404

	repo.On("FindByID", ctx, recipeID).Return(nil, ErrRecipeNotFound)

	err := svc.Rate(ctx, userID, recipeID, 3)

	assert.ErrorIs(t, err, ErrRecipeNotFound)
	repo.AssertNotCalled(t, "Rate", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// เทสของ attachMeta (ผ่าน List/GetByID) — พิสูจน์ว่า isFavorite และ rating.total
// ถูกคำนวณจริงจาก repository แทนที่จะ hardcode false/0 เหมือนเดิม
func TestService_List_AttachesFavoriteAndRatingMeta(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	query := GetRecipesQuery{
		Pagination: Pagination{Page: 1, Limit: 12},
		Sort:       DescendingSortDirection,
	}

	recipes := []Recipe{
		{ID: 1, Name: "Pad Thai"},
		{ID: 2, Name: "Tom Yum"},
	}

	repo.On("List", ctx, userID, query).Return(recipes, int64(2), nil)
	repo.On("FavoriteRecipeIDs", ctx, userID, []int{1, 2}).Return(map[int]bool{1: true}, nil)
	repo.On("RatingCounts", ctx, []int{1, 2}).Return(map[int]int64{2: 5}, nil)

	views, total, err := svc.List(ctx, userID, query)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, views, 2)

	assert.True(t, views[0].IsFavorite)
	assert.Equal(t, int64(0), views[0].RatingTotal)

	assert.False(t, views[1].IsFavorite)
	assert.Equal(t, int64(5), views[1].RatingTotal)
}

func TestService_GetByID_AttachesFavoriteAndRatingMeta(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()
	recipeID := 10
	recipe := &Recipe{ID: recipeID, Name: "Green Curry"}

	repo.On("FindByID", ctx, recipeID).Return(recipe, nil)
	repo.On("FavoriteRecipeIDs", ctx, userID, []int{recipeID}).Return(map[int]bool{recipeID: true}, nil)
	repo.On("RatingCounts", ctx, []int{recipeID}).Return(map[int]int64{recipeID: 3}, nil)

	view, err := svc.GetByID(ctx, userID, recipeID)

	require.NoError(t, err)
	assert.True(t, view.IsFavorite)
	assert.Equal(t, int64(3), view.RatingTotal)
	assert.Equal(t, "Green Curry", view.Name)
}

func TestService_GetByID_RecipeNotFound(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	ctx := context.Background()
	userID := uuid.New()

	repo.On("FindByID", ctx, 999).Return(nil, ErrRecipeNotFound)

	_, err := svc.GetByID(ctx, userID, 999)

	assert.ErrorIs(t, err, ErrRecipeNotFound)
}
