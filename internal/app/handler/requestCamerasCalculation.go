package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"awesomeProject/internal/app/ds"

	"github.com/gin-gonic/gin"
)

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
		simplifiedRequests = append(simplifiedRequests, gin.H{
			"id":           req.ID,
			"project_name": req.ProjectName,
			"status":       req.Status,
			"created_at":   req.CreatedAt,
			"formed_at":    req.FormedAt,
			"completed_at": req.CompletedAt,
		})
	}

	ctx.JSON(http.StatusOK, simplifiedRequests)
}

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

	ctx.JSON(http.StatusOK, gin.H{
		"id":           request.ID,
		"project_name": request.ProjectName,
		"status":       request.Status,
		"created_at":   request.CreatedAt,
		"formed_at":    request.FormedAt,
		"completed_at": request.CompletedAt,
		"calculations": simplifiedCalcs,
	})
}

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

func (h *Handler) DraftRequestCamerasCalculationInfoAPI(ctx *gin.Context) {
	draft, calcs, err := h.Repository.GetDraftRequestCamerasCalculationInfo()
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

func (h *Handler) FormRequestCamerasCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.FormDraft(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Заявка сформирована",
	})
}

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
