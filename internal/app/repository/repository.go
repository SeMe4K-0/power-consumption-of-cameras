package repository

import (
	"Lab1/internal/app/models"
	"fmt"
)

type Repository struct {
	cameras       []models.Camera
	orders        map[int]models.EnergyOrder
	orderServices []models.OrderService
	calculations  []models.EnergyCalculation
}

func NewRepository() (*Repository, error) {
	cameras := []models.Camera{
		{
			ID:          1,
			Name:        "Hikvision DS-2CD2043G0-I",
			Description: "4-мегапиксельная купольная камера для помещений. Инфракрасная подсветка до 30м, вандалоустойчивый корпус.",
			Power:       7.5,
			Type:        "Внутренняя",
			Resolution:  "4MP (2560x1440)",
			NightVision: true,
			ImageKey:    "http://127.0.0.1:9000/camers/DS-2CD2043G0-I.jpg",
		},
		{
			ID:          2,
			Name:        "Dahua IPC-HFW2831T-ZS",
			Description: "8-мегапиксельная уличная купольная камера с зумом. Защита IP67, работа при -40°C до +60°C.",
			Power:       12.0,
			Type:        "Уличная",
			Resolution:  "8MP (3840x2160)",
			NightVision: true,
			ImageKey:    "http://127.0.0.1:9000/camers/IPC-HFW2831T-ZS.png",
		},
		{
			ID:          3,
			Name:        "Axis M3045-V",
			Description: "Компактная сетевая камера для помещений. HD качество, встроенный микрофон, простая установка.",
			Power:       5.2,
			Type:        "Внутренняя",
			Resolution:  "2MP (1920x1080)",
			NightVision: false,
			ImageKey:    "http://127.0.0.1:9000/camers/AXIS%20M3045-V.jpg",
		},
		{
			ID:          4,
			Name:        "TP-Link Tapo C310",
			Description: "Поворотная уличная камера 3MP. Панорамирование 360°, цветное ночное видение, детектор движения.",
			Power:       9.8,
			Type:        "Уличная",
			Resolution:  "3MP (2304x1296)",
			NightVision: true,
			ImageKey:    "http://127.0.0.1:9000/camers/Tapo%20C310.jpg",
		},
		{
			ID:          5,
			Name:        "Reolink RLC-811A",
			Description: "Уличная камера 8MP с мощным зумом. Автоматическое слежение, цветное ночное видение, аудио.",
			Power:       15.3,
			Type:        "Уличная",
			Resolution:  "8MP (3840x2160)",
			NightVision: true,
			ImageKey:    "http://127.0.0.1:9000/camers/Reolink%20RLC-811A.jpg",
		},
	}

	// Создаем заявку для работы с вложенными камерами
	orders := map[int]models.EnergyOrder{
		1: {
			ID:          1,
			Tariff:      5.2, // руб/кВт·ч
			TotalPower:  12.0,
			DailyEnergy: 0.288, // 12 * 24 / 1000
			MonthlyCost: 44.93, // 0.288 * 30 * 5.2
			Status:      "В работе",
			ClientName:  "ООО 'ТехноСервис'",
			ProjectName: "Система видеонаблюдения офиса",
			Cameras: []models.OrderCamera{
				{
					Camera:   cameras[0], // Hikvision DS-2CD2043G0-I
					Quantity: 1,
					Hours:    24,
					Comment:  "Камера для входа",
				},
				{
					Camera:   cameras[1], // Dahua IPC-HFW2831T-ZS
					Quantity: 1,
					Hours:    24,
					Comment:  "Уличная камера",
				},
				{
					Camera:   cameras[4], // Reolink RLC-811A
					Quantity: 1,
					Hours:    24,
					Comment:  "Камера с зумом",
				},
			},
		},
	}

	// Создаем связи заявка-услуга
	orderServices := []models.OrderService{
		{
			OrderID:  1,
			CameraID: 1,
			Quantity: 1,
			Order:    1,
			Comment:  "Камера для входа",
			Other:    "Установка на потолок",
		},
	}

	// Создаем расчеты электроэнергии на основе камер видеонаблюдения
	calculations := []models.EnergyCalculation{
		{
			ID:               1,
			Name:             "Hikvision DS-2CD2043G0-I",
			Power:            7.5,
			DailyConsumption: 0.18,  // 7.5 * 24 / 1000
			MonthlyCost:      28.08, // 0.18 * 30 * 5.2
			ImageKey:         "http://127.0.0.1:9000/camers/DS-2CD2043G0-I.jpg",
			Description:      "4-мегапиксельная купольная камера",
			Category:         "Видеонаблюдение",
		},
		{
			ID:               2,
			Name:             "Dahua IPC-HFW2831T-ZS",
			Power:            12.0,
			DailyConsumption: 0.288, // 12.0 * 24 / 1000
			MonthlyCost:      44.93, // 0.288 * 30 * 5.2
			ImageKey:         "http://127.0.0.1:9000/camers/IPC-HFW2831T-ZS.png",
			Description:      "8-мегапиксельная уличная камера",
			Category:         "Видеонаблюдение",
		},
		{
			ID:               3,
			Name:             "Reolink RLC-811A",
			Power:            15.3,
			DailyConsumption: 0.367, // 15.3 * 24 / 1000
			MonthlyCost:      57.25, // 0.367 * 30 * 5.2
			ImageKey:         "http://127.0.0.1:9000/camers/Reolink%20RLC-811A.jpg",
			Description:      "Уличная камера 8MP с зумом",
			Category:         "Видеонаблюдение",
		},
	}

	return &Repository{
		cameras:       cameras,
		orders:        orders,
		orderServices: orderServices,
		calculations:  calculations,
	}, nil
}

func (r *Repository) GetCameras() ([]models.Camera, error) {
	if len(r.cameras) == 0 {
		return nil, fmt.Errorf("список камер пуст")
	}
	return r.cameras, nil
}

func (r *Repository) GetCamerasBySearch(query string) ([]models.Camera, error) {
	var result []models.Camera
	for _, camera := range r.cameras {
		if contains(camera.Name, query) || contains(camera.Type, query) || contains(camera.Resolution, query) {
			result = append(result, camera)
		}
	}
	return result, nil
}

func (r *Repository) GetCameraByID(id int) (models.Camera, error) {
	for _, camera := range r.cameras {
		if camera.ID == id {
			return camera, nil
		}
	}
	return models.Camera{}, fmt.Errorf("камера с ID %d не найдена", id)
}

func (r *Repository) GetOrderByID(id int) (models.EnergyOrder, error) {
	order, exists := r.orders[id]
	if !exists {
		return models.EnergyOrder{}, fmt.Errorf("заявка с ID %d не найдена", id)
	}
	return order, nil
}

func (r *Repository) GetOrdersCount() int {
	return len(r.orders)
}

// Получить данные для формы заявки
func (r *Repository) GetOrderFormData(orderID int) (models.OrderFormData, error) {
	order, exists := r.orders[orderID]
	if !exists {
		return models.OrderFormData{}, fmt.Errorf("заявка с ID %d не найдена", orderID)
	}

	return models.OrderFormData{
		Order:            order,
		AvailableCameras: r.cameras,
		Calculations:     r.calculations,
	}, nil
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
