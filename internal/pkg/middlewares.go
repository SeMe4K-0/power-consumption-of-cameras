package pkg

import (
	"awesomeProject/internal/app/ds"
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const jwtPrefix = "Bearer "

func (a *Application) WithAuthCheck(requireModerator bool) func(ctx *gin.Context) {
	return func(gCtx *gin.Context) {
		// Пытаемся получить JWT из Authorization header
		jwtStr := gCtx.GetHeader("Authorization")

		if strings.HasPrefix(jwtStr, jwtPrefix) {
			jwtStr = jwtStr[len(jwtPrefix):]
		} else {
			// Если нет в header, пытаемся получить из Cookie
			cookieToken, err := gCtx.Cookie("jwt_token")
			if err != nil || cookieToken == "" {
				gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"status":      "error",
					"description": "authorization header or cookie missing",
				})
				return
			}
			jwtStr = cookieToken
		}

		ctx := context.Background()
		_, err := a.Redis.GetClient().Get(ctx, "blacklist:"+jwtStr).Result()
		if err == nil {
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":      "error",
				"description": "token has been revoked",
			})
			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.Config.JWT.Token), nil
		})
		if err != nil {
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":      "error",
				"description": "invalid or expired token",
			})
			log.Println(err)
			return
		}

		myClaims := token.Claims.(*ds.JWTClaims)

		gCtx.Set("user_id", myClaims.UserID)
		gCtx.Set("is_professor", myClaims.IsProfessor)

		if requireModerator && !myClaims.IsProfessor {
			gCtx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":      "error",
				"description": "moderator permissions required for this operation",
			})
			return
		}
	}
}

func (a *Application) WithOptionalAuthCheck() func(ctx *gin.Context) {
	return func(gCtx *gin.Context) {
		// Пытаемся получить JWT из Authorization header
		jwtStr := gCtx.GetHeader("Authorization")

		if strings.HasPrefix(jwtStr, jwtPrefix) {
			jwtStr = jwtStr[len(jwtPrefix):]
		} else {
			// Если нет в header, пытаемся получить из Cookie
			cookieToken, err := gCtx.Cookie("jwt_token")
			if err != nil || cookieToken == "" {
				return // Опциональная авторизация, просто выходим
			}
			jwtStr = cookieToken
		}

		ctx := context.Background()
		_, err := a.Redis.GetClient().Get(ctx, "blacklist:"+jwtStr).Result()
		if err == nil {
			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.Config.JWT.Token), nil
		})
		if err != nil {
			return
		}

		myClaims := token.Claims.(*ds.JWTClaims)

		gCtx.Set("user_id", myClaims.UserID)
		gCtx.Set("is_professor", myClaims.IsProfessor)
	}
}
