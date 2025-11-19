package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"awesomeProject/internal/app/ds"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCamerasAPI(ctx *gin.Context) {
	cameraName := ctx.Query("camera")

	cameras, err := h.Repository.GetCameras(cameraName)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, cameras)
}
func (h *Handler) GetCameraAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	camera, err := h.Repository.GetCamera(uint(id))
	if err != nil {
		if err.Error() == "camera not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, camera)
}
func (h *Handler) CreateCameraAPI(ctx *gin.Context) {
	type cameraInput struct {
		Name        string  `json:"name" binding:"required"`
		Power       float64 `json:"power" binding:"required"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
		NightVision bool    `json:"night_vision"`
	}

	var input cameraInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	camera := ds.Cameras{
		Name:        input.Name,
		Power:       input.Power,
		NightVision: input.NightVision,
	}
	if input.Description != "" {
		camera.Description = &input.Description
	}
	if input.Status != "" {
		camera.Status = input.Status
	} else {
		camera.Status = "active"
	}

	createdCamera, err := h.Repository.CreateCamera(camera)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, createdCamera)
}
func (h *Handler) UpdateCameraAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	type cameraUpdateRequest struct {
		Name        string  `json:"name"`
		Power       float64 `json:"power"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
		NightVision bool    `json:"night_vision"`
	}

	var req cameraUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	camera := ds.Cameras{
		Name:        req.Name,
		Power:       req.Power,
		NightVision: req.NightVision,
	}
	if req.Description != "" {
		camera.Description = &req.Description
	}
	if req.Status != "" {
		camera.Status = req.Status
	}

	err = h.Repository.UpdateCamera(uint(id), camera)
	if err != nil {
		if err.Error() == "camera not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Камера обновлена",
	})
}
func (h *Handler) DeleteCameraAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	camera, err := h.Repository.GetCamera(uint(id))
	if err != nil {
		if err.Error() == "camera not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	if camera.Image != nil {
		fileName := *camera.Image
		parts := strings.Split(fileName, "/")
		if len(parts) > 0 {
			fileName = parts[len(parts)-1]
		}
		if fileName != "" {
			err = h.Repository.DeleteFileFromMinIO(ctx.Request.Context(), fileName)
			if err != nil {
				log.Printf("Warning: failed to delete image from MinIO: %v", err)
			}
		}
	}

	err = h.Repository.DeleteCamera(uint(id))
	if err != nil {
		if err.Error() == "camera not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Камера удалена",
	})
}
func (h *Handler) UploadCameraImageAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	_, err = h.Repository.GetCamera(uint(id))
	if err != nil {
		if err.Error() == "camera not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	src, err := file.Open()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	defer src.Close()

	fileName := "camera_" + strconv.FormatUint(uint64(id), 10) + ".png"

	err = h.Repository.UploadFileToMinIO(
		context.Background(),
		fileName,
		src,
		file.Size,
		file.Header.Get("Content-Type"),
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	imagePath := "http://127.0.0.1:9000/cameras/" + fileName
	err = h.Repository.UpdateCameraImage(uint(id), imagePath)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Изображение загружено",
		"image_path": imagePath,
	})
}
func (h *Handler) AddCameraToRequestCamerasCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	type addCameraRequest struct {
		Power float64 `json:"power" binding:"required,gt=0"`
	}

	var req addCameraRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("мощность должна быть больше 0"))
		return
	}

	draft, _, err := h.Repository.GetDraftRequestCamerasCalculationInfo(userID.(uint))
	if err != nil {
		created, createErr := h.Repository.CreateRequestCamerasCalculationWithCamera(uint(id), req.Power, userID.(uint))
		if createErr != nil {
			if strings.Contains(createErr.Error(), "duplicate key value violates unique constraint") {
				h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("камера уже добавлена в заявку"))
				return
			}
			h.errorHandler(ctx, http.StatusInternalServerError, createErr)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"status":   "success",
			"message":  "Камера добавлена в новую заявку",
			"draft_id": created.ID,
		})
		return
	}

	if draft.Status != ds.RequestStatusDraft {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("заявка должна быть в статусе черновика"))
		return
	}

	if err := h.Repository.AddCamerasCalculationToRequest(draft.ID, uint(id), req.Power); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("камера уже добавлена в заявку"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Камера добавлена в заявку",
		"draft_id": draft.ID,
	})
}
