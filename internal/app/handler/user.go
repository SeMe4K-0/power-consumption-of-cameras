package handler

import (
	"awesomeProject/internal/app/ds"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResp struct {
	ExpiresIn   int64  `json:"expires_in"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type registerReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type errorResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

type successResponse struct {
	Message string `json:"message"`
}

func (h *Handler) Register(gCtx *gin.Context) {
	var user ds.User
	if err := gCtx.ShouldBindJSON(&user); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}
	createdUser, err := h.Repository.CreateUser(user)
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}
	gCtx.JSON(http.StatusCreated, createdUser)
}

func (h *Handler) Login(gCtx *gin.Context) {
	cfg := h.Config
	req := &loginReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByUsername(req.Username)
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"description": "invalid username or password",
		})
		return
	}

	if req.Username == user.Username && req.Password == user.Password {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWT.ExpiresIn)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "bitop-admin",
			},
			UserID:      user.ID,
			IsProfessor: user.IsProfessor,
		})

		if token == nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
			return
		}

		strToken, err := token.SignedString([]byte(cfg.JWT.Token))
		if err != nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
			return
		}

		gCtx.SetCookie(
			"jwt_token",
			strToken,
			int(cfg.JWT.ExpiresIn.Seconds()),
			"/",
			"",
			false,
			true,
		)

		gCtx.JSON(http.StatusOK, loginResp{
			ExpiresIn:   int64(cfg.JWT.ExpiresIn.Seconds()),
			AccessToken: strToken,
			TokenType:   "Bearer",
		})
		return
	}

	gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"status":      "error",
		"description": "invalid username or password",
	})
}

func (h *Handler) Logout(gCtx *gin.Context) {
	var token string

	authHeader := gCtx.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		cookieToken, err := gCtx.Cookie("jwt_token")
		if err != nil || cookieToken == "" {
			gCtx.JSON(http.StatusBadRequest, gin.H{
				"status":      "error",
				"description": "authorization header or cookie missing",
			})
			return
		}
		token = cookieToken
	}

	ctx := context.Background()
	err := h.Redis.GetClient().Set(ctx, "blacklist:"+token, "1", time.Hour*24).Err()
	if err != nil {
		gCtx.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"description": "failed to logout",
		})
		return
	}

	gCtx.SetCookie(
		"jwt_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	gCtx.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
}

func (h *Handler) GetMeAPI(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	user, err := h.Repository.GetUserByID(userID.(uint))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateMeAPI(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	var user ds.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.UpdateUser(userID.(uint), user); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	updatedUser, _ := h.Repository.GetUserByID(userID.(uint))
	ctx.JSON(http.StatusOK, updatedUser)
}
