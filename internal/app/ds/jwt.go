package ds

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID            uint `json:"user_id"`
	IsLeadingEngineer bool `json:"is_leading_engineer"`
}
