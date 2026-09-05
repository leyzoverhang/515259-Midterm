package recipe

import "time"

type CreateRecipeRequest struct {
	Name         string                      `json:"name" binding:"required"`
	Description  string                      `json:"description" binding:"required"`
	ImageURL     *string                     `json:"imageUrl" binding:"omitempty,url"`
	DifficultyID string                      `json:"difficultyId" binding:"required"`
	DurationID   string                      `json:"durationId" binding:"required"`
	Ingredients  *[]RecipeIngredientRequest  `json:"ingredients" binding:"required,dive"`
	Instructions *[]RecipeInstructionRequest `json:"instructions" binding:"required,dive"`
}

type RecipeIngredientRequest struct {
	Description string `json:"description" binding:"required"`
}

type RecipeInstructionRequest struct {
	Description string `json:"description" binding:"required"`
}

func (req CreateRecipeRequest) ToRecipe() Recipe {
	recipe := Recipe{
		Name:         req.Name,
		Description:  req.Description,
		ImageURL:     req.ImageURL,
		DifficultyID: req.DifficultyID,
		DurationID:   req.DurationID,
	}

	if req.Ingredients != nil {
		recipe.Ingredients = make([]RecipeIngredient, len(*req.Ingredients))
		for index, ingredient := range *req.Ingredients {
			recipe.Ingredients[index] = RecipeIngredient{Description: ingredient.Description}
		}
	}

	if req.Instructions != nil {
		recipe.Instructions = make([]RecipeInstruction, len(*req.Instructions))
		for index, instruction := range *req.Instructions {
			recipe.Instructions[index] = RecipeInstruction{Description: instruction.Description}
		}
	}

	return recipe
}

type ReplaceRecipeRequest = CreateRecipeRequest

type CreateRecipeResponse struct {
	ID int `json:"id"`
}

func NewCreateRecipeResponse(recipe Recipe) CreateRecipeResponse {
	return CreateRecipeResponse{ID: recipe.ID}
}

type SortDirection string

const (
	AscendingSortDirection  SortDirection = "ASC"
	DescendingSortDirection SortDirection = "DESC"
)

func (direction *SortDirection) Ensure() {
	if direction == nil {
		return
	}

	if *direction == "" {
		*direction = DescendingSortDirection
	}
}

type Pagination struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (pg *Pagination) Ensure() {
	if pg.Page <= 0 {
		pg.Page = 1

	}

	if pg.Limit <= 0 || pg.Limit > 100 {
		pg.Limit = 12

	}
}

type GetRecipesQuery struct {
	Pagination
	Name       string        `form:"name"`
	Difficulty string        `form:"difficulty"`
	Sort       SortDirection `form:"sort" binding:"omitempty,oneof=ASC DESC"`
	Favorite   *bool         `form:"favorite"`
}

func (query *GetRecipesQuery) Ensure() {
	query.Pagination.Ensure()
	query.Sort.Ensure()
}

// Response get recipes
type CreatorResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DifficultyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DurationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IngredientResponse struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type InstructionResponse struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type RatingResponse struct {
	Average float64 `json:"average"`
	Total   int64   `json:"total"`
}

type RecipeResponse struct {
	ID           int                   `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	ImageURL     *string               `json:"imageUrl"`
	Difficulty   DifficultyResponse    `json:"difficulty"`
	Duration     DurationResponse      `json:"duration"`
	Ingredients  []IngredientResponse  `json:"ingredients"`
	Instructions []InstructionResponse `json:"instructions"`
	Creator      CreatorResponse       `json:"creator"`
	IsFavorite   bool                  `json:"isFavorite"` // true ถ้า user ที่ล็อกอินกดโปรดสูตรนี้ไว้
	Rating       RatingResponse        `json:"rating"`
	CreatedAt    time.Time             `json:"createdAt"`
	UpdatedAt    time.Time             `json:"updatedAt"`
}

func NewRecipeResponse(view RecipeView) RecipeResponse {
	recipe := view.Recipe

	ingredients := make([]IngredientResponse, 0, len(recipe.Ingredients))
	for _, ingredient := range recipe.Ingredients {
		ingredients = append(ingredients, IngredientResponse{
			ID:          ingredient.ID,
			Description: ingredient.Description,
		})
	}

	instructions := make([]InstructionResponse, 0, len(recipe.Instructions))
	for _, instruction := range recipe.Instructions {
		instructions = append(instructions, InstructionResponse{
			ID:          instruction.ID,
			Description: instruction.Description,
		})
	}

	return RecipeResponse{
		ID:          recipe.ID,
		Name:        recipe.Name,
		Description: recipe.Description,
		ImageURL:    recipe.ImageURL,
		Difficulty: DifficultyResponse{
			ID:   recipe.Difficulty.ID,
			Name: recipe.Difficulty.Name,
		},
		Duration: DurationResponse{
			ID:   recipe.Duration.ID,
			Name: recipe.Duration.Name,
		},
		Ingredients:  ingredients,
		Instructions: instructions,
		Creator: CreatorResponse{
			ID:   recipe.Creator.ID.String(),
			Name: *recipe.Creator.Name,
		},

		// IsFavorite/RatingTotal มาจาก service.attachMeta (batch query แยกจาก List/FindByID)
		IsFavorite: view.IsFavorite,
		Rating: RatingResponse{
			Average: recipe.AverageRating,
			Total:   view.RatingTotal,
		},
		CreatedAt: recipe.CreatedAt,
		UpdatedAt: recipe.UpdatedAt,
	}
}

type ListResponse[T any] struct {
	Total   int64 `json:"total"`
	Results []T   `json:"results"`
}

type ListRecipesResponse ListResponse[RecipeResponse]

func NewListRecipesResponse(views []RecipeView, total int64) ListRecipesResponse {
	results := make([]RecipeResponse, 0, len(views))
	for _, view := range views {
		results = append(results, NewRecipeResponse(view))
	}

	return ListRecipesResponse{
		Total:   total,
		Results: results,
	}
}

type RateRecipeRequest struct {
	Score float64 `json:"rating" binding:"required,min=1,max=5"`
}
