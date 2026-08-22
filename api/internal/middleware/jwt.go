package middleware

import (
	"context"
	"net/http"
	"strings"
	"wongnok/internal/httputil"
	"wongnok/internal/reqctx"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	bearerPrefix = "Bearer "
)

type UserResolver interface {
	ResolveID(ctx context.Context, uid string) (uuid.UUID, error)
}

func JWT(verifier *oidc.IDTokenVerifier, userResolver UserResolver) gin.HandlerFunc {
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

		userID, err := userResolver.ResolveID(ctx.Request.Context(), idToken.Subject)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "user not found"})
			return
		}

		rctx := reqctx.WithSubject(ctx.Request.Context(), idToken.Subject)
		rctx = reqctx.WithUserID(rctx, userID)
		ctx.Request = ctx.Request.WithContext(rctx)

		ctx.Next()
	}
}
