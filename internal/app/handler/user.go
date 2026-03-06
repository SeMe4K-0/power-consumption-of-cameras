package handler

import (
	"awesomeProject/internal/app/ds"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

type login2FAResp struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type verifyCodeReq struct {
	SessionID string `json:"session_id"`
	Code      string `json:"code"`
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

	// Валидация email
	if user.Email == "" {
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("email is required"))
		return
	}

	if !h.EmailService.ValidateEmail(user.Email) {
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("invalid email format"))
		return
	}

	createdUser, err := h.Repository.CreateUser(user)
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}
	gCtx.JSON(http.StatusCreated, createdUser)
}

// generateVerificationCode генерирует 6-значный код
func generateVerificationCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += num.String()
	}
	return code, nil
}

func (h *Handler) Login(gCtx *gin.Context) {
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

	if req.Username != user.Username || req.Password != user.Password {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"description": "invalid username or password",
		})
		return
	}

	// Проверяем наличие email
	if user.Email == "" {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"description": "user email is not set",
		})
		return
	}

	// Генерируем код подтверждения
	code, err := generateVerificationCode()
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"description": "failed to generate verification code",
		})
		return
	}

	// Создаем session ID
	sessionID := uuid.New().String()

	// Сохраняем код в Redis на 5 минут
	ctx := context.Background()
	key := fmt.Sprintf("2fa_code:%s", sessionID)
	err = h.Redis.GetClient().Set(ctx, key, code, 5*time.Minute).Err()
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"description": "failed to save verification code",
		})
		return
	}

	// Сохраняем информацию о пользователе для сессии
	userKey := fmt.Sprintf("2fa_user:%s", sessionID)
	userData := fmt.Sprintf("%d:%t", user.ID, user.IsLeadingEngineer)
	err = h.Redis.GetClient().Set(ctx, userKey, userData, 5*time.Minute).Err()
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"description": "failed to save session data",
		})
		return
	}

	// Отправляем код на email
	err = h.EmailService.SendVerificationCode(user.Email, code)
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"description": "failed to send verification code",
		})
		return
	}

	gCtx.JSON(http.StatusOK, login2FAResp{
		SessionID: sessionID,
		Message:   "Verification code sent to your email",
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

func (h *Handler) VerifyCode(gCtx *gin.Context) {
	cfg := h.Config
	req := &verifyCodeReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"description": "invalid request",
		})
		return
	}

	if req.SessionID == "" || req.Code == "" {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"description": "session_id and code are required",
		})
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf("2fa_code:%s", req.SessionID)
	
	// Получаем код из Redis
	storedCode, err := h.Redis.GetClient().Get(ctx, key).Result()
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"description": "invalid or expired code",
		})
		return
	}

	// Проверяем код
	if storedCode != req.Code {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"description": "invalid code",
		})
		return
	}

	// Получаем информацию о пользователе
	userKey := fmt.Sprintf("2fa_user:%s", req.SessionID)
	userData, err := h.Redis.GetClient().Get(ctx, userKey).Result()
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"description": "session expired",
		})
		return
	}

	var userID uint
	var isLeadingEngineer bool
	_, err = fmt.Sscanf(userData, "%d:%t", &userID, &isLeadingEngineer)
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"description": "failed to parse user data",
		})
		return
	}

	// Удаляем использованные ключи из Redis
	h.Redis.GetClient().Del(ctx, key)
	h.Redis.GetClient().Del(ctx, userKey)

	// Создаем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWT.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "bitop-admin",
		},
		UserID:            userID,
		IsLeadingEngineer: isLeadingEngineer,
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
