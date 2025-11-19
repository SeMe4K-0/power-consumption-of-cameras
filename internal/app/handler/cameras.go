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

// GetCamerasAPI godoc
// @Summary Получение списка камер
// @Description Возвращает список всех камер с возможностью фильтрации по названию
// @Tags Cameras
// @Produce json
// @Param camera query string false "Название камеры для фильтрации"
// @Success 200 {array} ds.Cameras "Список камер"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /cameras [get]
func (h *Handler) GetCamerasAPI(ctx *gin.Context) {
	cameraName := ctx.Query("camera")

	cameras, err := h.Repository.GetCameras(cameraName)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, cameras)
}

// GetCameraAPI godoc
// @Summary Получение камеры по ID
// @Description Возвращает информацию о камере по её ID
// @Tags Cameras
// @Produce json
// @Param id path int true "ID камеры"
// @Success 200 {object} ds.Cameras "Информация о камере"
// @Failure 400 {object} errorResponse "Неверный формат ID"
// @Failure 404 {object} errorResponse "Камера не найдена"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /cameras/{id} [get]
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

// CreateCameraAPI godoc
// @Summary Создание новой камеры
// @Description Создает новую камеру в системе (требует права модератора)
// @Tags Cameras
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string,power=number,description=string,status=string,night_vision=boolean} true "Данные камеры"
// @Success 201 {object} ds.Cameras "Созданная камера"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /cameras [post]
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

// UpdateCameraAPI godoc
// @Summary Обновление камеры
// @Description Обновляет информацию о камере (требует права модератора)
// @Tags Cameras
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID камеры"
// @Param request body object{name=string,power=number,description=string,status=string,night_vision=boolean} true "Данные для обновления"
// @Success 200 {object} object{status=string,message=string} "Камера обновлена"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Камера не найдена"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /cameras/{id} [put]
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

// DeleteCameraAPI godoc
// @Summary Удаление камеры
// @Description Удаляет камеру из системы (требует права модератора)
// @Tags Cameras
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID камеры"
// @Success 200 {object} object{status=string,message=string} "Камера удалена"
// @Failure 400 {object} errorResponse "Неверный формат ID"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Камера не найдена"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /cameras/{id} [delete]
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

	// Удаляем изображение из MinIO, если оно существует
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

// UploadCameraImageAPI godoc
// @Summary Загрузка изображения камеры
// @Description Загружает изображение для камеры в MinIO (требует права модератора)
// @Tags Cameras
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID камеры"
// @Param image formData file true "Изображение камеры"
// @Success 200 {object} object{status=string,message=string,image_path=string} "Изображение загружено"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Камера не найдена"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /cameras/{id}/image [post]
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

	// Загружаем файл в MinIO
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

	// Сохраняем путь к файлу в базе данных
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

// AddCameraToRequestCamerasCalculationAPI godoc
// @Summary Добавление камеры в заявку
// @Description Добавляет камеру в черновик заявки или создает новую заявку
// @Tags Cameras
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID камеры"
// @Param request body object{power=number} true "Мощность камеры"
// @Success 200 {object} object{status=string,message=string,draft_id=int} "Камера добавлена в заявку"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /cameras/{id}/addCamera [post]
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
