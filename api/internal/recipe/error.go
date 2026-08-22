package recipe

import "errors"

var (
	ErrInvalidReferenceData = errors.New("invalid recipe reference data")
	ErrRecipeNotFound       = errors.New("recipe not found")
)
