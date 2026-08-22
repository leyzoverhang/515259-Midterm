package user

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
	FindByID(ctx context.Context, uid uuid.UUID) (*User, error)
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

// GetUser godoc
//
//	@Summary		ดึงข้อมูลจาก user แบบรายคน
//	@Description	ค้นหาข้อมูล User จาก UUID แล้วคืนข้อมูล User ที่เจอกลับมา
//	@Tags			users
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"User ID (UUID)"
//	@Success		200	{object}	user.UserResponse
//	@Failure		400	{object}	httputil.ErrorResponse
//	@Failure		404	{object}	httputil.ErrorResponse
//	@Failure		500	{object}	httputil.ErrorResponse
//	@Router			/users/{id} [get]
func (hdr *handler) GetUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id != "me" {
		ctx.AbortWithStatusJSON(http.StatusNotFound, httputil.ErrorResponse{Message: "not found"})
		return
	}

	uid, ok := reqctx.UserID(ctx.Request.Context())
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "user not found"})
		return
	}

	user, err := hdr.service.FindByID(ctx, uid)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, httputil.ErrorResponse{Message: err.Error()})

		case errors.Is(err, ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, httputil.ErrorResponse{Message: err.Error()})

		default:
			ctx.JSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: err.Error()})

		}

		return
	}

	ctx.JSON(http.StatusOK, NewUserResponse(*user))
}
