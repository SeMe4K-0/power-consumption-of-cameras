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

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создание нового пользователя в системе
// @Tags Profile
// @Accept json
// @Produce json
// @Param request body registerReq true "Данные для регистрации"
// @Success 201 {object} ds.User "Успешная регистрация"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Router /profile/register [post]
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

// Login godoc
// @Summary Авторизация пользователя
// @Description Авторизация пользователя по username и паролю
// @Tags Profile
// @Accept json
// @Produce json
// @Param request body loginReq true "Данные для входа"
// @Success 200 {object} loginResp "Успешная авторизация"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 401 {object} errorResponse "Неверный username или пароль"
// @Router /profile/login [post]
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

		// Устанавливаем JWT в Cookie
		gCtx.SetCookie(
			"jwt_token",                      // имя cookie
			strToken,                         // значение
			int(cfg.JWT.ExpiresIn.Seconds()), // maxAge в секундах
			"/",                              // path
			"",                               // domain
			false,                            // secure (false для http)
			true,                             // httpOnly
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

// Logout godoc
// @Summary Выход пользователя из системы
// @Description Добавление токена в черный список для выхода из системы
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} successResponse "Успешный выход"
// @Failure 400 {object} errorResponse "Отсутствует или неверный заголовок авторизации"
// @Failure 500 {object} errorResponse "Ошибка при выходе"
// @Router /profile/logout [post]
func (h *Handler) Logout(gCtx *gin.Context) {
	var token string

	// Пытаемся получить из Authorization header
	authHeader := gCtx.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Если нет в header, пытаемся получить из Cookie
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

	// Удаляем Cookie
	gCtx.SetCookie(
		"jwt_token", // имя cookie
		"",          // пустое значение
		-1,          // maxAge = -1 (удаляет cookie)
		"/",         // path
		"",          // domain
		false,       // secure
		true,        // httpOnly
	)

	gCtx.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
}

// GetMeAPI godoc
// @Summary Получить информацию о текущем пользователе
// @Description Возвращает информацию о залогиненном пользователе
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ds.User "Информация о пользователе"
// @Failure 401 {object} errorResponse "Пользователь не авторизован"
// @Failure 404 {object} errorResponse "Пользователь не найден"
// @Router /profile/me [get]
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

// UpdateMeAPI godoc
// @Summary Обновить информацию о текущем пользователе
// @Description Обновляет информацию о залогиненном пользователе
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ds.User true "Данные для обновления"
// @Success 200 {object} ds.User "Обновленная информация о пользователе"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 401 {object} errorResponse "Пользователь не авторизован"
// @Router /profile/me [put]
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
