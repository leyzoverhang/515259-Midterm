package recipe

import (
	"context"
	"errors"
	"net/http"
	"wongnok/internal/httputil"
	"wongnok/internal/reqctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, creatorID uuid.UUID, recipe Recipe) (*Recipe, error)
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

// Create godoc
//
//	@Summary		สร้างสูตรอาหาร
//	@Description	สร้างสูตรอาหารจากข้อมูลที่ระบุ โดยผู้ใช้ที่ยืนยันตัวตนแล้วจะเป็นผู้สร้างสูตร
//	@Tags			recipes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateRecipeRequest	true	"Recipe data"
//	@Success		201		{object}	CreateRecipeResponse
//	@Failure		400		{object}	httputil.ErrorResponse
//	@Failure		401		{object}	httputil.ErrorResponse
//	@Failure		500		{object}	httputil.ErrorResponse
//	@Router			/recipes [post]
func (hdr *handler) Create(ctx *gin.Context) {
	creatorID, ok := reqctx.UserID(ctx.Request.Context())
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "unauthorized"})
		return
	}

	var req CreateRecipeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid request"})
		return
	}

	recipe, err := hdr.service.Create(ctx.Request.Context(), creatorID, req.ToRecipe())
	if err != nil {
		if errors.Is(err, ErrInvalidReferenceData) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid request"})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "internal server error"})
		return
	}

	ctx.JSON(http.StatusCreated, NewCreateRecipeResponse(*recipe))
}
