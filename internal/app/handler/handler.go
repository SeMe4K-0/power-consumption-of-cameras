package handler

import (
	"awesomeProject/internal/app/config"
	"awesomeProject/internal/app/redis"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository        *repository.Repository
	Config            *config.Config
	AuthCheck         func(requireModerator bool) gin.HandlerFunc
	OptionalAuthCheck func() gin.HandlerFunc
	Redis             *redis.Client
}

func NewHandler(r *repository.Repository, cfg *config.Config, authCheck func(requireModerator bool) gin.HandlerFunc, optionalAuthCheck func() gin.HandlerFunc, redisClient *redis.Client) *Handler {
	return &Handler{
		Repository:        r,
		Config:            cfg,
		AuthCheck:         authCheck,
		OptionalAuthCheck: optionalAuthCheck,
		Redis:             redisClient,
	}
}

func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")
	{
		h.registerProfileEndpoints(api)
		h.registerPublicEndpoints(api)
		h.registerCameraEndpoints(api)
		h.registerRequestEndpoints(api)
	}
}

func (h *Handler) registerProfileEndpoints(api *gin.RouterGroup) {
	profile := api.Group("/profile")
	{
		profile.POST("/register", h.Register)
		profile.POST("/login", h.Login)
		profile.POST("/logout", h.Logout)

		profileProtected := profile.Group("")
		profileProtected.Use(h.AuthCheck(false))
		{
			profileProtected.GET("/me", h.GetMeAPI)
			profileProtected.PUT("/me", h.UpdateMeAPI)
		}
	}
}

func (h *Handler) registerPublicEndpoints(api *gin.RouterGroup) {
	api.GET("/cameras", h.GetCamerasAPI)
	api.GET("/cameras/:id", h.GetCameraAPI)

	camerasCart := api.Group("/request_cameras_calculations")
	camerasCart.Use(h.OptionalAuthCheck())
	{
		camerasCart.GET("/cart", h.DraftRequestCamerasCalculationInfoAPI)
	}
}

func (h *Handler) registerCameraEndpoints(api *gin.RouterGroup) {
	camerasModerator := api.Group("/cameras")
	camerasModerator.Use(h.AuthCheck(true))
	{
		camerasModerator.POST("", h.CreateCameraAPI)
		camerasModerator.PUT("/:id", h.UpdateCameraAPI)
		camerasModerator.DELETE("/:id", h.DeleteCameraAPI)
		camerasModerator.POST("/:id/image", h.UploadCameraImageAPI)
	}

	camerasUser := api.Group("/cameras")
	camerasUser.Use(h.AuthCheck(false))
	{
		camerasUser.POST("/:id/addCamera", h.AddCameraToRequestCamerasCalculationAPI)
	}
}

func (h *Handler) registerRequestEndpoints(api *gin.RouterGroup) {
	requests := api.Group("/request_cameras_calculations")
	requests.Use(h.AuthCheck(false))
	{
		requests.GET("", h.GetRequestCamerasCalculationsAPI)
		requests.GET("/:id", h.GetRequestCamerasCalculationAPI)
		requests.PUT("/:id", h.UpdateRequestCamerasCalculationAPI)
		requests.DELETE("/:id", h.DeleteRequestCamerasCalculationAPI)
		requests.PUT("/:id/form", h.FormRequestCamerasCalculationAPI)
	}

	requestsModerator := api.Group("/request_cameras_calculations")
	requestsModerator.Use(h.AuthCheck(true))
	{
		requestsModerator.PUT("/:id/complete", h.CompleteRequestCamerasCalculationAPI)
	}

	calculations := api.Group("/cameras_calculations")
	calculations.Use(h.AuthCheck(false))
	{
		calculations.PUT("/:id/camera/:cameraId", h.UpdateCalculationSpeedAPI)
		calculations.DELETE("/:id/camera/:cameraId", h.RemoveCameraFromRequestCamerasCalculationAPI)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
