package auth

import (
	"context"
	"errors"
	"net/http"
	"wongnok/internal/httputil"

	"github.com/gin-gonic/gin"
)

type Service interface {
	BuildLoginURL(ctx context.Context) (string, error)
	HandleCallback(ctx context.Context, code, state string) (string, error)
	ExchangeTicket(ctx context.Context, ticket string) (Credential, error)
	Logout(ctx context.Context, refreshToken string) error
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

// Callback godoc
//
//	@Summary		รับ callback จาก Keycloak
//	@Description	Keycloak redirect มาที่นี่เองอัตโนมัติ ไม่ใช่สิ่งที่ frontend เรียกตรงๆ
//	@Tags			auth
//	@Param			code	query		string	true	"Authorization code"
//	@Param			state	query		string	true	"State ที่ตรงกับตอน /auth/login"
//	@Success		302		{string}	string	"Redirect กลับ frontend พร้อม ticket"
//	@Failure		400		{object}	httputil.ErrorResponse
//	@Failure		401		{object}	httputil.ErrorResponse
//	@Failure		502		{object}	httputil.ErrorResponse
//	@Router			/auth/callback [get]
func (hdr *handler) Cabllback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")

	if code == "" || state == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: "missing code or state"})
		return
	}

	redirectURL, err := hdr.service.HandleCallback(ctx.Request.Context(), code, state)
	if err != nil {
		if errors.Is(err, ErrInvalidState) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "invalid or expired state"})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusBadGateway, httputil.ErrorResponse{Message: "cannot complete login"})
		return
	}

	ctx.Redirect(http.StatusFound, redirectURL)
}

// Exchange godoc
//
//	@Summary		แลก ticket เป็น credential จริง
//	@Description	Frontend เรียกผ่าน axios หลังถูก redirect กลับมาพร้อม ticket
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ExchangeRequest	true	"ticket"
//	@Success		200		{object}	Credential
//	@Failure		400		{object}	httputil.ErrorResponse
//	@Failure		401		{object}	httputil.ErrorResponse
//	@Router			/auth/exchange [post]
func (hdr *handler) Exchange(ctx *gin.Context) {
	var req ExchangeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: err.Error()})
		return
	}

	credential, err := hdr.service.ExchangeTicket(ctx.Request.Context(), req.Ticket)
	if err != nil {
		if errors.Is(err, ErrInvalidTicket) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "invalid or expired ticket"})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "cannot exchange ticket"})
		return
	}

	ctx.JSON(http.StatusOK, credential)
}

// Logout godoc
//
//	@Summary		ออกจากระบบ
//	@Description	Revoke refresh token ที่ Keycloak จริง
//	@Tags			auth
//	@Security		BearerAuth
//	@Accept			json
//	@Param			request	body	LogoutRequest	true	"refresh token"
//	@Success		204		"Logout สำเร็จ"
//	@Failure		400		{object}	httputil.ErrorResponse
//	@Failure		502		{object}	httputil.ErrorResponse
//	@Router			/auth/logout [post]
func (hdr *handler) Logout(ctx *gin.Context) {
	var req LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httputil.ErrorResponse{Message: err.Error()})
		return
	}

	if err := hdr.service.Logout(ctx.Request.Context(), req.RefreshToken); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, httputil.ErrorResponse{Message: "logout failed"})
		return
	}

	ctx.Status(http.StatusNoContent)
}
