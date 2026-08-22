package auth

import (
	"context"
	"net/http"
	"wongnok/internal/httputil"

	"github.com/gin-gonic/gin"
)

type Service interface {
	BuildLoginURL(ctx context.Context) (string, error)
}

type handler struct {
	service Service
}

func NewHandler(svc Service) *handler {
	return &handler{
		service: svc,
	}
}

// Login godoc
//
//	@Summary		เริ่ม login ผ่าน Keycloak
//	@Description	Redirect ไปหน้า login ของ Keycloak
//	@Tags			auth
//	@Success		302	{string}	string	"Redirect ไป Keycloak"
//	@Failure		500	{object}	httputil.ErrorResponse
//	@Router			/auth/login [get]
func (hdr *handler) Login(ctx *gin.Context) {
	authURL, err := hdr.service.BuildLoginURL(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "cannot start login"})
		return
	}

	ctx.Redirect(http.StatusFound, authURL)
}
