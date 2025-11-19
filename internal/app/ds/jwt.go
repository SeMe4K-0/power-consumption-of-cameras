package ds

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      uint `json:"user_id"`
	IsProfessor bool `json:"is_professor"`
}
