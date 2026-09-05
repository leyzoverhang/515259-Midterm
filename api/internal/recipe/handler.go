package recipe

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"wongnok/internal/httputil"
	"wongnok/internal/reqctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, creatorID uuid.UUID, recipe Recipe) (*Recipe, error)
	FindByID(ctx context.Context, id int) (*Recipe, error)
	List(ctx context.Context, userID uuid.UUID, query GetRecipesQuery) ([]RecipeView, int64, error)
	// GetByID คืน RecipeView (recipe + isFavorite/rating.total ของ userID ที่ร้องขอ)
	GetByID(ctx context.Context, userID uuid.UUID, id int) (*RecipeView, error)
	Favorite(ctx context.Context, userID uuid.UUID, recipeID int) error
	Unfavorite(ctx context.Context, userID uuid.UUID, recipeID int) error
	Rate(ctx context.Context, userID uuid.UUID, recipeID int, score float64) error
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

// GetRecipe godoc
//
//	@Summary		เรียกดูสูตรอาหารแบบรายรายการ
//	@Description	ค้นหาสูตรอาหารจาก id แล้วคืนข้อมูลสูตรอาหารที่เจอกลับมา
//	@Tags			recipes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Recipe ID"
//	@Success		200	{object}	RecipeResponse
//	@Failure		400	{object}	httputil.ErrorResponse
//	@Failure		401	{object}	httputil.ErrorResponse
//	@Failure		404	{object}	httputil.ErrorResponse
//	@Failure		500	{object}	httputil.ErrorResponse
//	@Router			/recipes/{id} [get]
func (hdr *handler) GetRecipe(ctx *gin.Context) {
	userID, ok := reqctx.UserID(ctx.Request.Context())
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "unauthorized"})
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid request"})
		return
	}

	view, err := hdr.service.GetByID(ctx.Request.Context(), userID, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecipeNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, httputil.ErrorResponse{Message: "recipe not found"})

		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "internal server error"})

		}
		return
	}

	ctx.JSON(http.StatusOK, NewRecipeResponse(*view))
}

// GetRecipes godoc
//
//	@Summary		เรียกดูสูตรอาหารทั้งหมด
//	@Description	ค้นหาสูตรอาหารทั้งหมด กรองด้วยชื่อ/ความยาก/รายการโปรด และเรียงตามเวลาที่สร้างได้
//	@Tags			recipes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name		query		string	false	"ชื่อของสูตรอาหาร"
//	@Param			difficulty	query		string	false	"Id ของความยากในการทำ"
//	@Param			favorite	query		bool	false	"กรองตามรายการโปรด: true = เฉพาะที่กดโปรด, false = เฉพาะที่ยังไม่กด, ไม่ส่ง = ทั้งหมด"
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
	// ดึง userID จากคนที่ล็อกอิน
	userID, ok := reqctx.UserID(ctx.Request.Context())
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "unauthorized"})
		return
	}

	var query GetRecipesQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, httputil.ErrorResponse{Message: err.Error()})
		return
	}

	// ส่ง userID ต่อให้ชั้น Service
	recipes, total, err := hdr.service.List(ctx.Request.Context(), userID, query)
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

// Favorite godoc
//
//	@Summary		เพิ่มสูตรอาหารในรายการโปรด
//	@Description	เพิ่มสูตรอาหารที่ระบุเข้ารายการโปรดของผู้ใช้ที่ล็อกอิน กดซ้ำได้โดยไม่เกิด error
//	@Tags			recipes
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Recipe ID"
//	@Success		204
//	@Failure		400	{object}	httputil.ErrorResponse
//	@Failure		401	{object}	httputil.ErrorResponse
//	@Failure		404	{object}	httputil.ErrorResponse
//	@Failure		500	{object}	httputil.ErrorResponse
//	@Router			/recipes/{id}/favorite [post]
func (hdr *handler) Favorite(ctx *gin.Context) {
	userID, ok := reqctx.UserID(ctx.Request.Context())
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "unauthorized"})
		return
	}

	recipeID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid recipe id"})
		return
	}

	if err := hdr.service.Favorite(ctx.Request.Context(), userID, recipeID); err != nil {
		if errors.Is(err, ErrRecipeNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, httputil.ErrorResponse{Message: "recipe not found"})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "internal server error"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Unfavorite godoc
//
//	@Summary		เอาสูตรอาหารออกจากรายการโปรด
//	@Description	ลบสูตรอาหารที่ระบุออกจากรายการโปรดของผู้ใช้ที่ล็อกอิน (soft delete)
//	@Tags			recipes
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Recipe ID"
//	@Success		204
//	@Failure		400	{object}	httputil.ErrorResponse
//	@Failure		401	{object}	httputil.ErrorResponse
//	@Failure		404	{object}	httputil.ErrorResponse
//	@Failure		500	{object}	httputil.ErrorResponse
//	@Router			/recipes/{id}/favorite [delete]
func (hdr *handler) Unfavorite(ctx *gin.Context) {
	userID, ok := reqctx.UserID(ctx.Request.Context())
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "unauthorized"})
		return
	}

	recipeID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid recipe id"})
		return
	}

	if err := hdr.service.Unfavorite(ctx.Request.Context(), userID, recipeID); err != nil {
		if errors.Is(err, ErrRecipeNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, httputil.ErrorResponse{Message: "recipe not found"})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "internal server error"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Rate godoc
//
//	@Summary		ให้คะแนนสูตรอาหาร
//	@Description	ให้คะแนนสูตรอาหารที่ระบุ (1-5) ให้ซ้ำได้ ค่าล่าสุดจะแทนที่ค่าเดิม แล้วคำนวณ averageRating ใหม่ทันที
//	@Tags			recipes
//	@Security		BearerAuth
//	@Accept			json
//	@Param			id		path	int					true	"Recipe ID"
//	@Param			request	body	RateRecipeRequest	true	"คะแนน 1-5"
//	@Success		204
//	@Failure		400	{object}	httputil.ErrorResponse
//	@Failure		401	{object}	httputil.ErrorResponse
//	@Failure		404	{object}	httputil.ErrorResponse
//	@Failure		500	{object}	httputil.ErrorResponse
//	@Router			/recipes/{id}/rating [post]
func (hdr *handler) Rate(ctx *gin.Context) {
	userID, ok := reqctx.UserID(ctx.Request.Context())
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "unauthorized"})
		return
	}

	recipeID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid recipe id"})
		return
	}

	var req RateRecipeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid request"})
		return
	}

	if err := hdr.service.Rate(ctx.Request.Context(), userID, recipeID, req.Score); err != nil {
		if errors.Is(err, ErrRecipeNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, httputil.ErrorResponse{Message: "recipe not found"})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "internal server error"})
		return
	}

	ctx.Status(http.StatusNoContent)
}
