package repository

import (
	"fmt"
	"it-maintenance-backend/internal/app/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

func (r *Repository) GetCameras() ([]models.Camera, error) {
	var cameras []models.Camera
	result := r.db.Table("public.cameras").Find(&cameras)
	fmt.Printf("DEBUG: Found %d cameras, error: %v\n", len(cameras), result.Error)
	return cameras, result.Error
}

func (r *Repository) GetCameraByID(id uint) (models.Camera, error) {
	var camera models.Camera
	result := r.db.First(&camera, id)
	return camera, result.Error
}

func (r *Repository) GetCamerasBySearch(query string) ([]models.Camera, error) {
	var cameras []models.Camera
	searchQuery := "%" + query + "%"
	result := r.db.Table("public.cameras").Where("(name ILIKE ? OR description ILIKE ? OR type ILIKE ?)", searchQuery, searchQuery, searchQuery).Find(&cameras)
	fmt.Printf("DEBUG: Search '%s' found %d cameras, error: %v\n", query, len(cameras), result.Error)
	return cameras, result.Error
}

func (r *Repository) HasDraftOrder(userID uint) bool {
	var count int64
	r.db.Model(&models.SurveillanceOrder{}).Where("creator_id = ? AND status = ?", userID, models.OrderStatusDraft).Count(&count)
	return count > 0
}

func (r *Repository) GetCurrentOrder(userID uint) (models.SurveillanceOrder, error) {
	var order models.SurveillanceOrder
	err := r.db.Preload("OrderCameras.Camera").Where("creator_id = ? AND status = ?", userID, models.OrderStatusDraft).First(&order).Error
	if err != nil {
		return models.SurveillanceOrder{}, err
	}
	return order, nil
}

func (r *Repository) CreateOrder(userID uint, clientName, projectName string) (models.SurveillanceOrder, error) {
	order := models.SurveillanceOrder{
		CreatorID:   userID,
		Status:      models.OrderStatusDraft,
		CreatedAt:   time.Now(),
		ClientName:  clientName,
		ProjectName: projectName,
	}
	result := r.db.Create(&order)
	return order, result.Error
}

func (r *Repository) AddCameraToOrder(orderID, cameraID uint, quantity int, comment, other string) error {
	var orderCamera models.OrderCamera
	err := r.db.Where("order_id = ? AND camera_id = ?", orderID, cameraID).First(&orderCamera).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == gorm.ErrRecordNotFound {
		orderCamera = models.OrderCamera{
			OrderID:  orderID,
			CameraID: cameraID,
			Quantity: quantity,
			Comment:  comment,
			Other:    other,
			OrderNum: 1, // По умолчанию 1, можно доработать логику порядка
			IsMain:   false,
		}
		result := r.db.Create(&orderCamera)
		return result.Error
	} else {
		orderCamera.Quantity += quantity
		result := r.db.Save(&orderCamera)
		return result.Error
	}
}

func (r *Repository) GetOrderFormData(userID uint) (models.OrderFormData, error) {
	var order models.SurveillanceOrder
	err := r.db.Preload("OrderCameras.Camera").Where("creator_id = ? AND status = ?", userID, models.OrderStatusDraft).First(&order).Error
	if err != nil {
		return models.OrderFormData{}, err
	}

	var availableCameras []models.Camera
	r.db.Where("status = ?", models.CameraStatusActive).Find(&availableCameras)

	return models.OrderFormData{
		Order:            order,
		AvailableCameras: availableCameras,
		OrderCameras:     order.OrderCameras,
	}, nil
}

func (r *Repository) GetOrdersCount(userID uint) int64 {
	var count int64
	r.db.Model(&models.SurveillanceOrder{}).Where("creator_id = ?", userID).Count(&count)
	return count
}

func (r *Repository) GetCurrentOrderServicesCount(userID uint) int64 {
	var count int64
	var order models.SurveillanceOrder
	err := r.db.Where("creator_id = ? AND status = ?", userID, models.OrderStatusDraft).First(&order).Error
	if err != nil {
		return 0 // Если заявки нет, возвращаем 0
	}

	r.db.Model(&models.OrderCamera{}).Where("order_id = ?", order.ID).Count(&count)
	return count
}

func (r *Repository) GetOrderServicesCount(orderID uint) int64 {
	var count int64
	r.db.Model(&models.OrderCamera{}).Where("order_id = ?", orderID).Count(&count)
	return count
}

func (r *Repository) CheckOrderAccess(orderID uint, userID uint) error {
	var order models.SurveillanceOrder
	err := r.db.Where("id = ? AND creator_id = ?", orderID, userID).First(&order).Error
	if err != nil {
		return err // Заявка не найдена или не принадлежит пользователю
	}

	servicesCount := r.GetOrderServicesCount(orderID)
	if servicesCount == 0 {
		return fmt.Errorf("заявка пуста")
	}

	return nil
}

func (r *Repository) GetFirstOrderID(userID uint) uint {
	var order models.SurveillanceOrder
	result := r.db.Where("creator_id = ? AND status = ?", userID, models.OrderStatusDraft).First(&order)
	if result.Error != nil {
		fmt.Printf("DEBUG: GetFirstOrderID error for user %d: %v\n", userID, result.Error)
		return 0
	}
	fmt.Printf("DEBUG: GetFirstOrderID found order %d for user %d\n", order.ID, userID)
	return order.ID
}

func (r *Repository) GetOrderByID(id uint) (models.SurveillanceOrder, error) {
	var order models.SurveillanceOrder
	result := r.db.First(&order, id)
	return order, result.Error
}

func (r *Repository) GetOrderCameras(orderID uint) ([]models.OrderCamera, error) {
	var orderCameras []models.OrderCamera

	rows, err := r.db.Raw(`
		SELECT oc.order_id, oc.camera_id, oc.quantity, oc.order_num, oc.is_main, oc.comment, oc.other,
		       c.id, c.name, c.description, c.status, c.image_url, c.price, c.power, c.type, c.resolution, c.night_vision, c.created_at, c.updated_at
		FROM order_cameras oc
		JOIN cameras c ON oc.camera_id = c.id
		WHERE oc.order_id = ?
		ORDER BY oc.order_num
	`, orderID).Rows()

	if err != nil {
		return orderCameras, err
	}
	defer rows.Close()

	for rows.Next() {
		var oc models.OrderCamera
		var c models.Camera

		err := rows.Scan(
			&oc.OrderID, &oc.CameraID, &oc.Quantity, &oc.OrderNum, &oc.IsMain, &oc.Comment, &oc.Other,
			&c.ID, &c.Name, &c.Description, &c.Status, &c.ImageURL, &c.Price, &c.Power, &c.Type, &c.Resolution, &c.NightVision, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return orderCameras, err
		}

		oc.Camera = c
		orderCameras = append(orderCameras, oc)
	}

	return orderCameras, nil
}

func (r *Repository) DeleteOrder(orderID uint) error {
	result := r.db.Exec("UPDATE surveillance_orders SET status = ?, completion_date = ? WHERE id = ?", models.OrderStatusDeleted, time.Now(), orderID)
	return result.Error
}
