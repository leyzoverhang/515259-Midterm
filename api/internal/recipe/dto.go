package recipe

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
