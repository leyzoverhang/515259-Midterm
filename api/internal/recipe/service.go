package recipe

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	HasActiveReferences(ctx context.Context, difficultyID, durationID string) (bool, error)
	Create(ctx context.Context, recipe Recipe) (*Recipe, error)
	FindByID(ctx context.Context, id int) (*Recipe, error)
	List(ctx context.Context, userID uuid.UUID, query GetRecipesQuery) ([]Recipe, int64, error)
	DifficultyExists(ctx context.Context, id string) (bool, error)
	Favorite(ctx context.Context, userID uuid.UUID, recipeID int) error
	Unfavorite(ctx context.Context, userID uuid.UUID, recipeID int) error
	Rate(ctx context.Context, userID uuid.UUID, recipeID int, score float64) error
	// FavoriteRecipeIDs/RatingCounts คำนวณ isFavorite / rating.total แบบ batch
	// (คนละ concern กับ List/FindByID เพื่อไม่ให้ query หลักพ่วงข้อมูลที่ขึ้นกับผู้เรียก)
	FavoriteRecipeIDs(ctx context.Context, userID uuid.UUID, recipeIDs []int) (map[int]bool, error)
	RatingCounts(ctx context.Context, recipeIDs []int) (map[int]int64, error)
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

func (svc *service) FindByID(ctx context.Context, id int) (*Recipe, error) {
	recipe, err := svc.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRecipeNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("find recipe: %w", err)
	}

	return recipe, nil
}

func (svc *service) List(ctx context.Context, userID uuid.UUID, query GetRecipesQuery) ([]RecipeView, int64, error) {
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

	recipes, total, err := svc.repository.List(ctx, userID, query)
	if err != nil {
		return nil, 0, fmt.Errorf("list recipes: %w", err)
	}

	views, err := svc.attachMeta(ctx, userID, recipes)
	if err != nil {
		return nil, 0, fmt.Errorf("list recipes: %w", err)
	}

	return views, total, nil
}

// attachMeta คำนวณ isFavorite และ rating.total ของแต่ละสูตรแบบ batch
// (2 query รวม ไม่ใช่ query ต่อสูตร) แล้วห่อ Recipe เดิมด้วยค่าที่ได้
func (svc *service) attachMeta(ctx context.Context, userID uuid.UUID, recipes []Recipe) ([]RecipeView, error) {
	ids := make([]int, len(recipes))
	for i, r := range recipes {
		ids[i] = r.ID
	}

	favorites, err := svc.repository.FavoriteRecipeIDs(ctx, userID, ids)
	if err != nil {
		return nil, err
	}

	counts, err := svc.repository.RatingCounts(ctx, ids)
	if err != nil {
		return nil, err
	}

	views := make([]RecipeView, len(recipes))
	for i, r := range recipes {
		views[i] = RecipeView{
			Recipe:      r,
			IsFavorite:  favorites[r.ID],
			RatingTotal: counts[r.ID],
		}
	}

	return views, nil
}

// GetByID คืน RecipeView (Recipe + isFavorite/rating.total ของ userID ที่ร้องขอ)
// ใช้กับ endpoint GET /recipes/:id เท่านั้น — ส่วน FindByID เดิมยังเก็บไว้ใช้
// เช็คว่าสูตรมีอยู่จริงภายใน Favorite/Unfavorite/Rate โดยไม่ต้องแบก query เพิ่ม
func (svc *service) GetByID(ctx context.Context, userID uuid.UUID, id int) (*RecipeView, error) {
	recipe, err := svc.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	views, err := svc.attachMeta(ctx, userID, []Recipe{*recipe})
	if err != nil {
		return nil, fmt.Errorf("get recipe: %w", err)
	}

	return &views[0], nil
}

// Favorite/Unfavorite/Rate ทั้งสามฟังก์ชันเช็คก่อนว่าสูตรมีอยู่จริง แล้วค่อยส่งต่อให้ repository
// เพื่อให้ handler แปลง ErrRecipeNotFound เป็น 404 ได้โดยไม่ต้องพึ่ง error จากชั้น DB
func (svc *service) Favorite(ctx context.Context, userID uuid.UUID, recipeID int) error {
	// เช็กว่าสูตรมีอยู่จริงไหม
	if _, err := svc.FindByID(ctx, recipeID); err != nil {
		return err // ถ้าไม่มีจะส่ง ErrRecipeNotFound กลับไปทันที
	}
	return svc.repository.Favorite(ctx, userID, recipeID)
}

func (svc *service) Unfavorite(ctx context.Context, userID uuid.UUID, recipeID int) error {
	if _, err := svc.FindByID(ctx, recipeID); err != nil {
		return err
	}
	return svc.repository.Unfavorite(ctx, userID, recipeID)
}

func (svc *service) Rate(ctx context.Context, userID uuid.UUID, recipeID int, score float64) error {
	if _, err := svc.FindByID(ctx, recipeID); err != nil {
		return err
	}
	return svc.repository.Rate(ctx, userID, recipeID, score)
}
