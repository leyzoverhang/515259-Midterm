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
	List(ctx context.Context, query GetRecipesQuery) ([]Recipe, int64, error)
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

// GetRecipes godoc
//
//	@Summary		เรียกดูสูตรอาหารทั้งหมด
//	@Description	ค้นหาสูตรอาหารทั้งหมด และสามารถกรองด้วย ชื่อ, ความยาก และเรียงตามเวลาที่สร้างได้
//	@Tags			recipes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name		query		string	false	"ชื่อของสูตรอาหาร"
//	@Param			difficulty	query		string	false	"Id ของความยากในการทำ"
//	@Param			sort		query		string	false	"เรียงลำดับตามเวลาที่สร้าง"	Enums(ASC, DESC)	default(DESC)
//	@Param			page		query		int		false	"หน้าที่ต้องการแสดง"				minimum(1)			default(1)
//	@Param			limit		query		int		false	"จำนวนรายการต่อหน้า"			minimum(1)			maximum(100)	default(12)
//	@Success		200			{object}	ListRecipesResponse
//	@Failure		400			{object}	httputil.ErrorResponse
//	@Failure		401			{object}	httputil.ErrorResponse
//	@Failure		404			{object}	httputil.ErrorResponse
//	@Failure		500			{object}	httputil.ErrorResponse
//	@Router			/recipes [get]
func (hdr *handler) GetRecipes(ctx *gin.Context) {
	var query GetRecipesQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, httputil.ErrorResponse{Message: err.Error()})
		return
	}

	recipes, total, err := hdr.service.List(ctx.Request.Context(), query)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidReferenceData):
			ctx.JSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid request"})

		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "internal server error"})

		}
		return
	}

	ctx.JSON(http.StatusOK, NewListRecipesResponse(recipes, total))
}
