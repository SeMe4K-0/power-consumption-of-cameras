package ds

type CamerasCalculation struct {
	ID                            uint     `gorm:"primaryKey;autoIncrement"`
	RequestCamerasCalculationID uint     `gorm:"not null;uniqueIndex:idx_request_cameras"`
	CamerasID                    uint     `gorm:"not null;uniqueIndex:idx_request_cameras"`
	Power                         float64  `gorm:"type:numeric;not null"`
	MonthlyCost                   *float64 `gorm:"type:numeric"`

	Request  RequestCamerasCalculation `gorm:"foreignKey:RequestCamerasCalculationID"`
	Cameras Cameras                    `gorm:"foreignKey:CamerasID"`
}
