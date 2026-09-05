package recipe

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}

func (repo *repository) HasActiveReferences(ctx context.Context, difficultyID, durationID string) (bool, error) {
	var difficultyCount int64
	if err := repo.db.WithContext(ctx).Model(&Difficulty{}).Where("id = ?", difficultyID).Count(&difficultyCount).Error; err != nil {
		return false, err
	}

	if difficultyCount == 0 {
		return false, nil
	}

	var durationCount int64
	if err := repo.db.WithContext(ctx).Model(&Duration{}).Where("id = ?", durationID).Count(&durationCount).Error; err != nil {
		return false, err
	}

	return durationCount > 0, nil
}

func (repo *repository) Create(ctx context.Context, recipe Recipe) (*Recipe, error) {
	if err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(&recipe).Error; err != nil {
			return err
		}

		for index := range recipe.Ingredients {
			recipe.Ingredients[index].RecipeID = recipe.ID
		}
		if len(recipe.Ingredients) > 0 {
			if err := tx.Omit(clause.Associations).Create(&recipe.Ingredients).Error; err != nil {
				return err
			}
		}

		for index := range recipe.Instructions {
			recipe.Instructions[index].RecipeID = recipe.ID
		}
		if len(recipe.Instructions) > 0 {
			if err := tx.Omit(clause.Associations).Create(&recipe.Instructions).Error; err != nil {
				return err
			}
		}

		return nil

	}); err != nil {
		return nil, err

	}

	return &recipe, nil
}

func (repo *repository) FindByID(ctx context.Context, id int) (*Recipe, error) {
	db := repo.db.WithContext(ctx).Preload("Difficulty").Preload("Duration").Preload("Creator").Preload("Ingredients").Preload("Instructions")

	var recipe Recipe
	if err := db.First(&recipe, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecipeNotFound
		}

		return nil, err
	}

	return &recipe, nil
}

// List รับ userID เพิ่มเพื่อกรองรายการโปรด (query.Favorite) เฉพาะของผู้ใช้ที่ล็อกอิน
func (repo *repository) List(ctx context.Context, userID uuid.UUID, query GetRecipesQuery) ([]Recipe, int64, error) {
	db := repo.db.WithContext(ctx).Model(&Recipe{})

	if query.Name != "" {
		db = db.Where("name ILIKE ?", ("%" + query.Name + "%"))
	}
	if query.Difficulty != "" {
		db = db.Where("difficulty_id = ?", query.Difficulty)
	}

	// กับดัก T3: ใช้ Subquery เพื่อกรอง Favorite ไม่ใช่ JOIN
	if query.Favorite != nil {
		subQuery := repo.db.WithContext(ctx).Model(&UserFavorite{}).Select("recipe_id").Where("user_id = ?", userID)
		if *query.Favorite {
			db = db.Where("recipes.id IN (?)", subQuery)
		} else {
			db = db.Where("recipes.id NOT IN (?)", subQuery)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count recipes: %w", err)
	}

	order := "created_at DESC"
	if query.Sort != "" {
		order = fmt.Sprintf("created_at %s", query.Sort)
	}

	// Preload
	db = db.Preload("Difficulty").Preload("Duration").Preload("Creator").Preload("Ingredients").Preload("Instructions")

	// Order
	db = db.Order(order)

	// Pagination
	db = db.Offset((query.Page - 1) * query.Limit).Limit(query.Limit)

	// Find
	var recipes []Recipe
	if err := db.Find(&recipes).Error; err != nil {
		return nil, 0, fmt.Errorf("list recipes: %w", err)
	}

	return recipes, total, nil
}

func (repo *repository) DifficultyExists(ctx context.Context, id string) (bool, error) {
	var count int64

	if err := repo.db.WithContext(ctx).Model(&Difficulty{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check difficulty %q: %w", id, err)
	}

	return count > 0, nil
}

func (repo *repository) Favorite(ctx context.Context, userID uuid.UUID, recipeID int) error {
	// กับดัก T1: ใช้ ON CONFLICT เพื่อคืนชีพแถวข้อมูล
	return repo.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "recipe_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"deleted_at": nil,
			"updated_at": time.Now(),
		}),
	}).Create(&UserFavorite{
		UserID:   userID,
		RecipeID: recipeID,
	}).Error
}

func (repo *repository) Unfavorite(ctx context.Context, userID uuid.UUID, recipeID int) error {
	// การลบแบบ Soft Delete อาศัย gorm.DeletedAt ใน Struct
	return repo.db.WithContext(ctx).
		Where("user_id = ? AND recipe_id = ?", userID, recipeID).
		Delete(&UserFavorite{}).Error
}

func (repo *repository) Rate(ctx context.Context, userID uuid.UUID, recipeID int, score float64) error {
	// กับดัก T4: ต้องทำทั้ง 3 ขั้นตอนใน Transaction เดียว
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1. บันทึกหรืออัปเดตคะแนนของ User คนนี้
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "recipe_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"score":      score,
				"deleted_at": nil,
				"updated_at": time.Now(),
			}),
		}).Create(&RecipeRating{
			UserID:   userID,
			RecipeID: recipeID,
			Score:    score,
		}).Error
		if err != nil {
			return err
		}

		// 2. คำนวณ Average Rating ใหม่ของสูตรนี้ (กับดัก T5: COALESCE)
		var avgScore float64
		err = tx.Model(&RecipeRating{}).
			Select("COALESCE(AVG(score), 0)").
			Where("recipe_id = ? AND deleted_at IS NULL", recipeID).
			Scan(&avgScore).Error
		if err != nil {
			return err
		}

		// 3. นำค่า Average ที่คำนวณได้ไปอัปเดตลงในตาราง Recipes
		return tx.Model(&Recipe{}).
			Where("id = ?", recipeID).
			Update("average_rating", avgScore).Error
	})
}

func (repo *repository) FavoriteRecipeIDs(ctx context.Context, userID uuid.UUID, recipeIDs []int) (map[int]bool, error) {
	result := make(map[int]bool, len(recipeIDs))
	if len(recipeIDs) == 0 {
		return result, nil
	}

	var favorited []int
	if err := repo.db.WithContext(ctx).
		Model(&UserFavorite{}).
		Select("recipe_id").
		Where("user_id = ? AND recipe_id IN ? AND deleted_at IS NULL", userID, recipeIDs).
		Find(&favorited).Error; err != nil {
		return nil, fmt.Errorf("favorite recipe ids: %w", err)
	}

	for _, id := range favorited {
		result[id] = true
	}

	return result, nil
}

func (repo *repository) RatingCounts(ctx context.Context, recipeIDs []int) (map[int]int64, error) {
	result := make(map[int]int64, len(recipeIDs))
	if len(recipeIDs) == 0 {
		return result, nil
	}

	type ratingCountRow struct {
		RecipeID int
		Total    int64
	}

	var rows []ratingCountRow
	if err := repo.db.WithContext(ctx).
		Model(&RecipeRating{}).
		Select("recipe_id, COUNT(*) AS total").
		Where("recipe_id IN ? AND deleted_at IS NULL", recipeIDs).
		Group("recipe_id").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("rating counts: %w", err)
	}

	for _, row := range rows {
		result[row.RecipeID] = row.Total
	}

	return result, nil
}
