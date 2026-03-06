package models

import (
	"time"
)

const (
	CameraStatusActive   = "действует"
	CameraStatusInactive = "удален"
)

const (
	OrderStatusDraft     = "черновик"
	OrderStatusDeleted   = "удален"
	OrderStatusFormed    = "сформирован"
	OrderStatusCompleted = "завершен"
	OrderStatusRejected  = "отклонен"
)

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"uniqueIndex" json:"username"`                // Логин
	Password    string    `gorm:"not null" json:"-"`                          // Пароль (скрыт в JSON)
	IsModerator bool      `gorm:"not null;default:false" json:"is_moderator"` // Признак модератора
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`     // Активен ли пользователь
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Camera struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"not null;default:'действует'" json:"status"`  // "действует" / "удален"
	ImageURL    *string   `gorm:"column:image_url;type:text" json:"image_url"` // URL к изображению (Nullable)
	Price       float64   `gorm:"not null;default:0" json:"price"`
	Power       float64   `gorm:"not null;default:0" json:"power"`                                // Мощность в Ваттах
	Type        string    `gorm:"not null" json:"type"`                                           // Тип: уличная/внутренняя
	Resolution  string    `gorm:"not null" json:"resolution"`                                     // Разрешение
	NightVision bool      `gorm:"column:night_vision;not null;default:false" json:"night_vision"` // Ночное видение
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Camera) TableName() string {
	return "cameras"
}

type SurveillanceOrder struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Status          string     `gorm:"not null" json:"status"` // Статус заявки
	CreatedAt       time.Time  `gorm:"not null" json:"created_at"`
	CreatorID       uint       `gorm:"not null" json:"creator_id"` // Создатель
	FormationDate   *time.Time `json:"formation_date"`             // Дата формирования (2 действия создателя)
	CompletionDate  *time.Time `json:"completion_date"`            // Дата завершения (2 действия модератора)
	ModeratorID     *uint      `json:"moderator_id"`               // Модератор
	ClientName      string     `gorm:"column:client_name;not null" json:"client_name"`
	ProjectName     string     `gorm:"column:project_name;not null" json:"project_name"`
	CalculatedField float64    `gorm:"default:0" json:"calculated_field"` // Доп. поле, рассчитываемое при завершении

	Creator      User          `gorm:"foreignKey:CreatorID" json:"creator"`
	Moderator    *User         `gorm:"foreignKey:ModeratorID" json:"moderator"`
	OrderCameras []OrderCamera `gorm:"foreignKey:OrderID" json:"order_cameras"`
}

type OrderCamera struct {
	OrderID  uint   `gorm:"not null" json:"order_id"`              // ID заявки
	CameraID uint   `gorm:"not null" json:"camera_id"`             // ID камеры
	Quantity int    `gorm:"not null;default:1" json:"quantity"`    // Количество
	OrderNum int    `gorm:"not null;default:1" json:"order_num"`   // Порядок
	IsMain   bool   `gorm:"not null;default:false" json:"is_main"` // Главный/дополнительный
	Comment  string `gorm:"type:text" json:"comment"`              // Комментарий
	Other    string `gorm:"type:text" json:"other"`                // Другое

	Order  SurveillanceOrder `gorm:"foreignKey:OrderID" json:"order"`
	Camera Camera            `gorm:"foreignKey:CameraID" json:"camera"`
}

type OrderFormData struct {
	Order            SurveillanceOrder `json:"order"`
	AvailableCameras []Camera          `json:"available_cameras"`
	OrderCameras     []OrderCamera     `json:"order_cameras"`
}
