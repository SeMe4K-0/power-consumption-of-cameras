package ds

type Cameras struct {
	ID          uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string  `gorm:"type:varchar(50);not null" json:"name"`
	Power       float64 `gorm:"type:numeric;not null" json:"power"`
	Image       *string `gorm:"type:varchar(50);default:null" json:"image,omitempty"`
	Description *string `gorm:"type:text;default:null" json:"description,omitempty"`
	Status      string  `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	NightVision bool    `gorm:"type:boolean;default:false" json:"night_vision"`
	IsDeleted   bool    `gorm:"type:boolean;not null;default:false" json:"-"`
}
