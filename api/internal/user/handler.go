package user

import (
	"context"
	"errors"
	"net/http"
	"wongnok/internal/httputil"

	"github.com/gin-gonic/gin"
)

type Service interface {
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, user User) (*User, error)
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (hdr *handler) GetUser(ctx *gin.Context) {
	uid := ctx.Param("id")

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

func (hdr *handler) CreateUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid request body"})
		return
	}

	user, err := hdr.service.Create(ctx, req.ToUser())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httputil.ErrorResponse{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, NewUserResponse(*user))
}
