package repository

import (
	"awesomeProject/internal/app/ds"
)

func (r *Repository) RemoveCalculationFromRequest(requestID, cameraID uint) (int64, error) {
	result := r.db.Where("surveillance_orders_id = ? AND cameras_id = ?", requestID, cameraID).Delete(&ds.CamerasCalculation{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *Repository) UpdateCalculationValue(requestID uint, cameraID uint, power float64) error {
	return r.db.Model(&ds.CamerasCalculation{}).Where("surveillance_orders_id = ? AND cameras_id = ?", requestID, cameraID).Update("power", power).Error
}
