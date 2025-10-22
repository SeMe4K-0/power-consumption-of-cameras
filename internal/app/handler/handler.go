package handler

import (
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")
	{
		profile := api.Group("/profile")
		{
			profile.POST("/register", h.RegisterUserAPI)
			profile.POST("/login", h.LoginAPI)
			profile.POST("/logout", h.LogoutAPI)
			profile.GET("/me", h.GetMeAPI)
			profile.PUT("/me", h.UpdateMeAPI)
		}

		cameras := api.Group("/cameras")
		{
			cameras.GET("", h.GetCamerasAPI)
			cameras.GET("/:id", h.GetCameraAPI)
			cameras.POST("", h.CreateCameraAPI)
			cameras.PUT("/:id", h.UpdateCameraAPI)
			cameras.DELETE("/:id", h.DeleteCameraAPI)
			cameras.POST("/:id/image", h.UploadCameraImageAPI)
			cameras.POST("/:id/addCamera", h.AddCameraToRequestCamerasCalculationAPI)
		}

		requests := api.Group("/requestcamerascalculations")
		{
			requests.GET("/cart", h.DraftRequestCamerasCalculationInfoAPI)
			requests.GET("", h.GetRequestCamerasCalculationsAPI)
			requests.GET("/:id", h.GetRequestCamerasCalculationAPI)
			requests.PUT("/:id", h.UpdateRequestCamerasCalculationAPI)
			requests.DELETE("/:id", h.DeleteRequestCamerasCalculationAPI)
			requests.PUT("/:id/form", h.FormRequestCamerasCalculationAPI)
			requests.PUT("/:id/complete", h.CompleteRequestCamerasCalculationAPI)
		}

		calculations := api.Group("/camerascalculations")
		{
			calculations.PUT("/:id/camera/:cameraId", h.UpdateCalculationSpeedAPI)
			calculations.DELETE("/:id/camera/:cameraId", h.RemoveCameraFromRequestCamerasCalculationAPI)
		}
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
