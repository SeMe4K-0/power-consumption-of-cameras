package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"awesomeProject/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// GetRequestCamerasCalculationsAPI godoc
// @Summary Получение списка заявок
// @Description Возвращает список заявок пользователя с фильтрацией по статусу и дате
// @Tags RequestCamerasCalculations
// @Produce json
// @Security BearerAuth
// @Param status query string false "Статус заявки"
// @Param start_date query string false "Начальная дата (формат: 2006-01-02)"
// @Param end_date query string false "Конечная дата (формат: 2006-01-02)"
// @Success 200 {array} object "Список заявок"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /request_cameras_calculations [get]
func (h *Handler) GetRequestCamerasCalculationsAPI(ctx *gin.Context) {
	var status *ds.RequestStatus
	var startDate, endDate *time.Time

	if statusStr := ctx.Query("status"); statusStr != "" {
		requestStatus := ds.RequestStatus(statusStr)
		status = &requestStatus
	}

	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	requests, err := h.Repository.GetRequestCamerasCalculations(status, startDate, endDate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	var simplifiedRequests []gin.H
	for _, req := range requests {
		// Получаем количество результатов (камер с рассчитанным monthly_cost)
		resultsCount, _ := h.Repository.GetResultsCountForRequest(req.ID)

		// Формируем логины пользователей
		creatorLogin := ""
		if req.Creator.Username != "" {
			creatorLogin = req.Creator.Username
		}
		moderatorLogin := ""
		if req.Moderator != nil && req.Moderator.Username != "" {
			moderatorLogin = req.Moderator.Username
		}

		simplifiedRequests = append(simplifiedRequests, gin.H{
			"id":            req.ID,
			"project_name":  req.ProjectName,
			"status":        req.Status,
			"created_at":    req.CreatedAt,
			"formed_at":     req.FormedAt,
			"completed_at":  req.CompletedAt,
			"creator":       creatorLogin,
			"moderator":     moderatorLogin,
			"results_count": resultsCount,
		})
	}

	ctx.JSON(http.StatusOK, simplifiedRequests)
}

// GetRequestCamerasCalculationAPI godoc
// @Summary Получение заявки по ID
// @Description Возвращает детальную информацию о заявке с расчетами
// @Tags RequestCamerasCalculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} object "Детальная информация о заявке"
// @Failure 400 {object} errorResponse "Неверный формат ID"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Заявка не найдена"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /request_cameras_calculations/{id} [get]
func (h *Handler) GetRequestCamerasCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	request, calcs, err := h.Repository.GetRequestWithCalculations(uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	var simplifiedCalcs []gin.H
	for _, calc := range calcs {
		simplifiedCalcs = append(simplifiedCalcs, gin.H{
			"id":           calc.ID,
			"camera_id":    calc.CamerasID,
			"camera_name":  calc.Cameras.Name,
			"camera_power": calc.Cameras.Power,
			"camera_image": calc.Cameras.Image,
			"power":        calc.Power,
			"monthly_cost": calc.MonthlyCost,
		})
	}

	// Формируем информацию о пользователях
	creatorLogin := ""
	if request.Creator.Username != "" {
		creatorLogin = request.Creator.Username
	}
	moderatorLogin := ""
	if request.Moderator != nil && request.Moderator.Username != "" {
		moderatorLogin = request.Moderator.Username
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":           request.ID,
		"project_name": request.ProjectName,
		"status":       request.Status,
		"created_at":   request.CreatedAt,
		"formed_at":    request.FormedAt,
		"completed_at": request.CompletedAt,
		"creator":      creatorLogin,
		"moderator":    moderatorLogin,
		"calculations": simplifiedCalcs,
	})
}

// UpdateRequestCamerasCalculationAPI godoc
// @Summary Обновление заявки
// @Description Обновляет информацию о заявке (только для создателя)
// @Tags RequestCamerasCalculations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param request body ds.RequestCamerasCalculation true "Данные для обновления"
// @Success 200 {object} object{status=string,message=string} "Заявка обновлена"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Заявка не найдена"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /request_cameras_calculations/{id} [put]
func (h *Handler) UpdateRequestCamerasCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var request ds.RequestCamerasCalculation
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.UpdateRequestCamerasCalculation(uint(id), request)
	if err != nil {
		if err.Error() == "record not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Заявка обновлена",
	})
}

// DeleteRequestCamerasCalculationAPI godoc
// @Summary Удаление заявки
// @Description Удаляет заявку из системы (только для создателя)
// @Tags RequestCamerasCalculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} object{status=string,message=string} "Заявка удалена"
// @Failure 400 {object} errorResponse "Неверный формат ID"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Заявка не найдена"
// @Failure 500 {object} errorResponse "Внутренняя ошибка сервера"
// @Router /request_cameras_calculations/{id} [delete]
func (h *Handler) DeleteRequestCamerasCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	rowsAffected, err := h.Repository.DeleteRequestCamerasCalculation(uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	if rowsAffected == 0 {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("заявка не найдена"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Заявка удалена",
	})
}

// UpdateRequestStatusAPI godoc
// @Summary Обновление статуса заявки
// @Description Обновляет статус заявки
// @Tags RequestCamerasCalculations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param request body object{status=string,moderator_id=int} true "Новый статус и ID модератора"
// @Success 200 {object} object{status=string,message=string} "Статус заявки обновлен"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Заявка не найдена"
// @Router /request_cameras_calculations/{id}/status [put]
func (h *Handler) UpdateRequestStatusAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req struct {
		Status      ds.RequestStatus `json:"status" binding:"required"`
		ModeratorID *uint            `json:"moderator_id,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.UpdateRequestStatus(uint(id), req.Status, req.ModeratorID)
	if err != nil {
		if err.Error() == "record not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Статус заявки обновлен",
	})
}

// DraftRequestCamerasCalculationInfoAPI godoc
// @Summary Информация о черновике заявки
// @Description Возвращает информацию о текущем черновике заявки пользователя
// @Tags RequestCamerasCalculations
// @Produce json
// @Success 200 {object} object{draft_id=int,cameras_cnt=int} "Информация о черновике"
// @Router /request_cameras_calculations/cart [get]
func (h *Handler) DraftRequestCamerasCalculationInfoAPI(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		userID = uint(1) // Для гостя используем ID 1
	}

	draft, calcs, err := h.Repository.GetDraftRequestCamerasCalculationInfo(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"draft_id":    0,
			"cameras_cnt": 0,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"draft_id":    draft.ID,
		"cameras_cnt": len(calcs),
	})
}

// FormRequestCamerasCalculationAPI godoc
// @Summary Формирование заявки
// @Description Переводит черновик заявки в статус "сформирован"
// @Tags RequestCamerasCalculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} object{status=string,message=string} "Заявка сформирована"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Router /request_cameras_calculations/{id}/form [put]
func (h *Handler) FormRequestCamerasCalculationAPI(ctx *gin.Context) {
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

	if err := h.Repository.FormDraft(uint(id), userID.(uint)); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Заявка сформирована",
	})
}

// CompleteRequestCamerasCalculationAPI godoc
// @Summary Завершение заявки модератором
// @Description Завершает или отклоняет заявку (требует права модератора)
// @Tags RequestCamerasCalculations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param request body object{approve=boolean} false "Одобрить заявку (по умолчанию true)"
// @Success 200 {object} object{status=string,message=string} "Заявка обработана"
// @Failure 400 {object} errorResponse "Неверный формат запроса"
// @Failure 403 {object} errorResponse "Доступ запрещен (требуется модератор)"
// @Router /request_cameras_calculations/{id}/complete [put]
func (h *Handler) CompleteRequestCamerasCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req struct {
		Approve bool `json:"approve"`
	}

	// Если JSON не предоставлен, используем значение по умолчанию
	if err := ctx.ShouldBindJSON(&req); err != nil {
		req.Approve = true // По умолчанию одобряем заявку
	}

	if err := h.Repository.CompleteRequest(uint(id), req.Approve); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	message := "Заявка отклонена"
	if req.Approve {
		message = "Заявка одобрена и обработана"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": message,
	})
}
