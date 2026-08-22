package middleware

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}
