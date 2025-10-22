package repository

import (
	"awesomeProject/internal/app/ds"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) GetRequestCamerasCalculations(status *ds.RequestStatus, startDate, endDate *time.Time) ([]ds.RequestCamerasCalculation, error) {
	var requests []ds.RequestCamerasCalculation
	query := r.db.Where("status != ? AND status != ?", ds.RequestStatusDeleted, ds.RequestStatusDraft)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if startDate != nil {
		query = query.Where("formed_at >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("formed_at <= ?", *endDate)
	}

	err := query.Preload("Creator").Preload("Moderator").Find(&requests).Error
	return requests, err
}

func (r *Repository) GetRequestCamerasCalculation(id uint) (ds.RequestCamerasCalculation, error) {
	var request ds.RequestCamerasCalculation
	err := r.db.Preload("Creator").Preload("Moderator").Where("id = ? AND status != ?", id, ds.RequestStatusDeleted).First(&request).Error
	return request, err
}

func (r *Repository) GetRequestWithCalculations(id uint) (ds.RequestCamerasCalculation, []ds.CamerasCalculation, error) {
	req, err := r.GetRequestCamerasCalculation(id)
	if err != nil {
		return ds.RequestCamerasCalculation{}, nil, err
	}
	var calcs []ds.CamerasCalculation
	err = r.db.Preload("Cameras").Where("request_cameras_calculation_id = ?", id).Find(&calcs).Error
	if err != nil {
		return ds.RequestCamerasCalculation{}, nil, err
	}
	return req, calcs, nil
}

func (r *Repository) UpdateRequestCamerasCalculation(id uint, request ds.RequestCamerasCalculation) error {
	var existingRequest ds.RequestCamerasCalculation
	err := r.db.Where("id = ? AND status != ?", id, ds.RequestStatusDeleted).First(&existingRequest).Error
	if err != nil {
		return err
	}

	return r.db.Model(&existingRequest).Updates(request).Error
}

func (r *Repository) UpdateRequestStatus(id uint, newStatus ds.RequestStatus, moderatorID *uint) error {
	var request ds.RequestCamerasCalculation
	err := r.db.Where("id = ? AND status != ?", id, ds.RequestStatusDeleted).First(&request).Error
	if err != nil {
		return err
	}

	if !r.isValidStatusTransition(request.Status, newStatus) {
		return fmt.Errorf("недопустимый переход статуса с %s на %s", request.Status, newStatus)
	}

	updates := map[string]interface{}{
		"status": newStatus,
	}

	switch newStatus {
	case ds.RequestStatusFormed:
		updates["formed_at"] = time.Now()
	case ds.RequestStatusCompleted, ds.RequestStatusRejected:
		updates["completed_at"] = time.Now()
		updates["moderator_id"] = *moderatorID
	}

	return r.db.Model(&ds.RequestCamerasCalculation{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) isValidStatusTransition(current, new ds.RequestStatus) bool {
	validTransitions := map[ds.RequestStatus][]ds.RequestStatus{
		ds.RequestStatusDraft:     {ds.RequestStatusDeleted, ds.RequestStatusFormed},
		ds.RequestStatusFormed:    {ds.RequestStatusCompleted, ds.RequestStatusRejected},
		ds.RequestStatusCompleted: {},
		ds.RequestStatusRejected:  {},
		ds.RequestStatusDeleted:   {},
	}

	allowedStatuses, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == new {
			return true
		}
	}
	return false
}

func (r *Repository) GetDraftRequestCamerasCalculationInfo() (ds.RequestCamerasCalculation, []ds.CamerasCalculation, error) {
	creatorID := ds.GetCreatorID()

	requestCamerasCalculation := ds.RequestCamerasCalculation{}
	err := r.db.Preload("Creator").Preload("Moderator").Where("creator_id = ? AND status = ?", creatorID, ds.RequestStatusDraft).First(&requestCamerasCalculation).Error
	if err != nil {
		return ds.RequestCamerasCalculation{}, nil, err
	}

	var camerasCalculations []ds.CamerasCalculation
	err = r.db.Preload("Cameras").Where("request_cameras_calculation_id = ?", requestCamerasCalculation.ID).Find(&camerasCalculations).Error
	if err != nil {
		return ds.RequestCamerasCalculation{}, nil, err
	}

	return requestCamerasCalculation, camerasCalculations, nil
}

func (r *Repository) calculateMonthlyCost(power float64) float64 {
	// Расчет месячной стоимости: мощность (Вт) * 24 часа * 30 дней * 5 руб/кВтч / 1000
	return (power * 24 * 30 * 5.0) / 1000
}

func (r *Repository) calculateMonthlyCostsForRequest(requestID uint) error {
	var calculations []ds.CamerasCalculation
	err := r.db.Preload("Cameras").Where("request_cameras_calculation_id = ?", requestID).Find(&calculations).Error
	if err != nil {
		return err
	}

	for _, calc := range calculations {
		cost := r.calculateMonthlyCost(calc.Power)
		err = r.db.Model(&ds.CamerasCalculation{}).Where("id = ?", calc.ID).Update("monthly_cost", &cost).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) AddCamerasCalculationToRequest(requestID uint, cameraID uint, power float64) error {
	camerasCalculation := ds.CamerasCalculation{
		RequestCamerasCalculationID: requestID,
		CamerasID:                   cameraID,
		Power:                       power,
		MonthlyCost:                 nil,
	}
	err := r.db.Create(&camerasCalculation).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteRequestCamerasCalculation(id uint) (int64, error) {
	var existingRequest ds.RequestCamerasCalculation
	err := r.db.Where("id = ? AND status != ?", id, ds.RequestStatusDeleted).First(&existingRequest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("record not found")
		}
		return 0, err
	}

	result := r.db.Model(&existingRequest).Update("status", ds.RequestStatusDeleted)
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (r *Repository) FormDraft(id uint) error {
	draft, calcs, err := r.GetDraftRequestCamerasCalculationInfo()
	if err != nil || draft.ID != id {
		return fmt.Errorf("доступен только черновик текущего пользователя")
	}
	if len(calcs) == 0 {
		return fmt.Errorf("заявка пуста")
	}
	newStatus := ds.RequestStatusFormed
	return r.UpdateRequestStatus(id, newStatus, nil)
}

func (r *Repository) CompleteRequest(id uint, approve bool) error {
	status := ds.RequestStatusRejected
	if approve {
		status = ds.RequestStatusCompleted
	}
	mod := ds.GetCreatorID()

	err := r.UpdateRequestStatus(id, status, &mod)
	if err != nil {
		return err
	}

	if approve && status == ds.RequestStatusCompleted {
		err = r.calculateMonthlyCostsForRequest(id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) CreateRequestCamerasCalculationWithCamera(cameraID uint, power float64) (ds.RequestCamerasCalculation, error) {
	requestCamerasCalculation := ds.RequestCamerasCalculation{
		ProjectName: "Система видеонаблюдения",
		Status:      ds.RequestStatusDraft,
		CreatorID:   ds.GetCreatorID(),
	}
	err := r.db.Create(&requestCamerasCalculation).Error
	if err != nil {
		return ds.RequestCamerasCalculation{}, err
	}

	err = r.AddCamerasCalculationToRequest(requestCamerasCalculation.ID, cameraID, power)
	if err != nil {
		return ds.RequestCamerasCalculation{}, err
	}

	return requestCamerasCalculation, nil
}
