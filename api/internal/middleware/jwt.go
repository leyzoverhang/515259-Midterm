package middleware

import (
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

const (
	bearerPrefix = "Bearer "
)

func JWT(verifier *oidc.IDTokenVerifier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		rawToken := strings.TrimPrefix(authHeader, bearerPrefix)

		idToken, err := verifier.Verify(ctx.Request.Context(), rawToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		ctx.Set("subject", idToken.Subject)
		ctx.Next()
	}
}
