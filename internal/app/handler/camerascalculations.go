package handler

import (
	"awesomeProject/internal/app/ds"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UpdateCalculationSpeedAPI godoc
// @Summary Обновление мощности камеры в расчете
// @Description Обновляет мощность камеры в черновике заявки
// @Tags CamerasCalculations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param cameraId path int true "ID камеры"
// @Param request body object{power=number} true "Новая мощность"
// @Success 200 {object} object{status=string,message=string} "Мощность обновлена"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Заявка не найдена"
// @Router /cameras_calculations/{id}/camera/{cameraId} [put]
func (h *Handler) UpdateCalculationSpeedAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	cameraIdStr := ctx.Param("cameraId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	cameraID, err := strconv.ParseUint(cameraIdStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	request, err := h.Repository.GetRequestCamerasCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if request.Status != ds.RequestStatusDraft {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("заявка должна быть в статусе черновика"))
		return
	}

	var body struct {
		Power float64 `json:"power" binding:"required,gt=0"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("мощность должна быть больше 0"))
		return
	}

	if err := h.Repository.UpdateCalculationValue(uint(id), uint(cameraID), body.Power); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Мощность обновлена",
	})
}

// RemoveCameraFromRequestCamerasCalculationAPI godoc
// @Summary Удаление камеры из заявки
// @Description Удаляет камеру из черновика заявки
// @Tags CamerasCalculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param cameraId path int true "ID камеры"
// @Success 200 {object} object{status=string,message=string} "Камера удалена из заявки"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Камера не найдена в заявке"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /cameras_calculations/{id}/camera/{cameraId} [delete]
func (h *Handler) RemoveCameraFromRequestCamerasCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	cameraIdStr := ctx.Param("cameraId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	cameraID, err := strconv.ParseUint(cameraIdStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	request, err := h.Repository.GetRequestCamerasCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if request.Status != ds.RequestStatusDraft {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("заявка должна быть в статусе черновика"))
		return
	}

	rowsAffected, err := h.Repository.RemoveCalculationFromRequest(uint(id), uint(cameraID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if rowsAffected == 0 {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("камера не найдена в заявке"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Камера удалена из заявки",
	})
}
