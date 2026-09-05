package recipe

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"wongnok/internal/user"
)

type Difficulty struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type Duration struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type Recipe struct {
	ID            int `gorm:"primaryKey"`
	Name          string
	Description   string
	ImageURL      *string
	DifficultyID  string
	Difficulty    Difficulty
	DurationID    string
	Duration      Duration
	AverageRating float64
	CreatorID     uuid.UUID
	Creator       user.User `gorm:"foreignKey:CreatorID;references:ID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt
	Ingredients   []RecipeIngredient  `gorm:"foreignKey:RecipeID"`
	Instructions  []RecipeInstruction `gorm:"foreignKey:RecipeID"`
}

type RecipeIngredient struct {
	ID          int `gorm:"primaryKey"`
	RecipeID    int
	Recipe      Recipe `gorm:"foreignKey:RecipeID;references:ID"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt
}

type RecipeInstruction struct {
	ID          int `gorm:"primaryKey"`
	RecipeID    int
	Recipe      Recipe `gorm:"foreignKey:RecipeID;references:ID"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt
}

type UserFavorite struct {
	UserID    uuid.UUID      `gorm:"primaryKey"`
	RecipeID  int            `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type RecipeRating struct {
	UserID    uuid.UUID      `gorm:"primaryKey"`
	RecipeID  int            `gorm:"primaryKey"`
	Score     float64        `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// RecipeView bundles a Recipe with per-viewer metadata that depends on who
// is asking (favorite status, total number of ratings) rather than on the
// recipe row itself.
type RecipeView struct {
	Recipe
	IsFavorite  bool
	RatingTotal int64
}

