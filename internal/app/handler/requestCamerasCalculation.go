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

	userID, userExists := ctx.Get("user_id")
	isProfessor, isProfessorExists := ctx.Get("is_professor")

	var userIDPtr *uint
	var isProfessorBool bool

	if userExists {
		uid := userID.(uint)
		userIDPtr = &uid
	}
	if isProfessorExists {
		isProfessorBool = isProfessor.(bool)
	}

	requests, err := h.Repository.GetRequestCamerasCalculations(status, startDate, endDate, userIDPtr, isProfessorBool)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	var simplifiedRequests []gin.H
	for _, req := range requests {
		resultsCount, _ := h.Repository.GetResultsCountForRequest(req.ID)

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

	existingRequest, err := h.Repository.GetRequestCamerasCalculation(uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists || existingRequest.CreatorID != userID.(uint) {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("доступ запрещен: вы можете обновлять только свои заявки"))
		return
	}

	if request.Status == ds.RequestStatusCompleted || request.Status == ds.RequestStatusRejected {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("доступ запрещен: создатель не может завершать заявки"))
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
	userID, exists := ctx.Get("user_id")
	if !exists {
		userID = uint(1)
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

	if err := ctx.ShouldBindJSON(&req); err != nil {
		req.Approve = true
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	if err := h.Repository.CompleteRequest(uint(id), req.Approve, userID.(uint)); err != nil {
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
