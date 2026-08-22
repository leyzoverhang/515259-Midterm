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
