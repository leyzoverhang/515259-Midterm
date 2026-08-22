package recipe

import (
	"context"
	"fmt"

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

func (repo *repository) List(ctx context.Context, query GetRecipesQuery) ([]Recipe, int64, error) {
	db := repo.db.WithContext(ctx).Model(&Recipe{})

	if query.Name != "" {
		db = db.Where("name ILIKE ?", ("%" + query.Name + "%"))
	}
	if query.Difficulty != "" {
		db = db.Where("difficulty_id = ?", query.Difficulty)
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
	db = db.Order((query.Page - 1) * query.Limit).Limit(query.Limit)

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
