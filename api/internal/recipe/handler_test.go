package recipe

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"wongnok/internal/reqctx"
	"wongnok/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestContext สร้าง gin.Context เปล่า ๆ สำหรับยิงเข้า handler โดยตรง
// (ข้าม router/middleware จริง) — withUserID ควบคุมว่าจะมี userID ใน context ไหม
// (จำลองผลลัพธ์ของ middleware.JWT)
func newTestContext(t *testing.T, method, path string, body []byte, params gin.Params, userID *uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	if userID != nil {
		req = req.WithContext(reqctx.WithUserID(req.Context(), *userID))
	}

	ctx.Request = req
	ctx.Params = params

	return ctx, w
}

func TestHandler_Favorite_Success_Returns204(t *testing.T) {
	svc := new(MockService)
	hdr := NewHandler(svc)

	userID := uuid.New()
	svc.On("Favorite", mock.Anything, userID, 42).Return(nil)

	ctx, w := newTestContext(t, http.MethodPost, "/api/v1/recipes/42/favorite", nil,
		gin.Params{{Key: "id", Value: "42"}}, &userID)

	hdr.Favorite(ctx)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_Favorite_Unauthorized_Returns401(t *testing.T) {
	svc := new(MockService)
	hdr := NewHandler(svc)

	ctx, w := newTestContext(t, http.MethodPost, "/api/v1/recipes/42/favorite", nil,
		gin.Params{{Key: "id", Value: "42"}}, nil)

	hdr.Favorite(ctx)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	svc.AssertNotCalled(t, "Favorite")
}

func TestHandler_Favorite_InvalidID_Returns400(t *testing.T) {
	svc := new(MockService)
	hdr := NewHandler(svc)

	userID := uuid.New()
	ctx, w := newTestContext(t, http.MethodPost, "/api/v1/recipes/abc/favorite", nil,
		gin.Params{{Key: "id", Value: "abc"}}, &userID)

	hdr.Favorite(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Favorite")
}

func TestHandler_Favorite_RecipeNotFound_Returns404(t *testing.T) {
	svc := new(MockService)
	hdr := NewHandler(svc)

	userID := uuid.New()
	svc.On("Favorite", mock.Anything, userID, 999).Return(ErrRecipeNotFound)

	ctx, w := newTestContext(t, http.MethodPost, "/api/v1/recipes/999/favorite", nil,
		gin.Params{{Key: "id", Value: "999"}}, &userID)

	hdr.Favorite(ctx)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_Rate_InvalidScore_Returns400(t *testing.T) {
	svc := new(MockService)
	hdr := NewHandler(svc)

	userID := uuid.New()
	body, err := json.Marshal(map[string]any{"rating": 9}) // นอกช่วง 1-5
	require.NoError(t, err)

	ctx, w := newTestContext(t, http.MethodPost, "/api/v1/recipes/7/rating", body,
		gin.Params{{Key: "id", Value: "7"}}, &userID)

	hdr.Rate(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Rate")
}

func TestHandler_Rate_Success_Returns204(t *testing.T) {
	svc := new(MockService)
	hdr := NewHandler(svc)

	userID := uuid.New()
	svc.On("Rate", mock.Anything, userID, 7, 4.0).Return(nil)

	body, err := json.Marshal(map[string]any{"rating": 4})
	require.NoError(t, err)

	ctx, w := newTestContext(t, http.MethodPost, "/api/v1/recipes/7/rating", body,
		gin.Params{{Key: "id", Value: "7"}}, &userID)

	hdr.Rate(ctx)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestHandler_GetRecipe_ReturnsFavoriteAndRatingFromView ยืนยันว่า isFavorite และ
// rating.total ที่ service คำนวณมา (ไม่ใช่ค่า hardcode) ถูก serialize ออกไปใน JSON จริง
func TestHandler_GetRecipe_ReturnsFavoriteAndRatingFromView(t *testing.T) {
	svc := new(MockService)
	hdr := NewHandler(svc)

	userID := uuid.New()
	creatorName := "Chef Somchai"
	view := &RecipeView{
		Recipe: Recipe{
			ID:     10,
			Name:   "Green Curry",
			Creator: user.User{ID: userID, Name: &creatorName},
		},
		IsFavorite:  true,
		RatingTotal: 4,
	}
	svc.On("GetByID", mock.Anything, userID, 10).Return(view, nil)

	ctx, w := newTestContext(t, http.MethodGet, "/api/v1/recipes/10", nil,
		gin.Params{{Key: "id", Value: "10"}}, &userID)

	hdr.GetRecipe(ctx)

	require.Equal(t, http.StatusOK, w.Code)

	var resp RecipeResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.IsFavorite)
	assert.Equal(t, int64(4), resp.Rating.Total)
}
