package middleware

import (
	"crypto/subtle"
	"net/http"
	"wongnok/internal/httputil"

	"github.com/gin-gonic/gin"
)

const (
	basicAuthUser = "admin"
	basicAuthPass = "secret"
)

func BasicAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, pass, ok := ctx.Request.BasicAuth()

		validUser := subtle.ConstantTimeCompare([]byte(user), []byte(basicAuthUser)) == 1
		validPassword := subtle.ConstantTimeCompare([]byte(pass), []byte(basicAuthPass)) == 1

		if !ok || !validUser || !validPassword {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{
				Message: http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		ctx.Next()
	}
}
