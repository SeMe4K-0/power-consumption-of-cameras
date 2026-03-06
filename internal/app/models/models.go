package models

type OldCamera struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Power       float64 `json:"power"`        // Мощность в Ваттах
	Type        string  `json:"type"`         // Тип: уличная/внутренняя
	Resolution  string  `json:"resolution"`   // Разрешение
	NightVision bool    `json:"night_vision"` // Ночное видение
	ImageKey    string  `json:"image_key"`    // Ключ изображения в Minio
}

type OldOrderCamera struct {
	Camera   OldCamera `json:"camera"`   // Основная информация о камере
	Quantity int       `json:"quantity"` // Количество камер
	Hours    int       `json:"hours"`    // Часы работы в сутки
	Comment  string    `json:"comment"`  // Комментарий
}

type EnergyOrder struct {
	ID          int              `json:"id"`
	TotalPower  float64          `json:"total_power"`  // Общая мощность (Вт)
	DailyEnergy float64          `json:"daily_energy"` // Суточное потребление (кВт·ч)
	MonthlyCost float64          `json:"monthly_cost"` // Стоимость за месяц (руб)
	Tariff      float64          `json:"tariff"`       // Тариф на электроэнергию
	Status      string           `json:"status"`
	ClientName  string           `json:"client_name"`  // Имя клиента
	ProjectName string           `json:"project_name"` // Название проекта
	Cameras     []OldOrderCamera `json:"cameras"`      // Вложенный массив камер
}

type EnergyCalculation struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`              // Название устройства
	Power            float64 `json:"power"`             // Мощность в Ваттах
	DailyConsumption float64 `json:"daily_consumption"` // Суточное потребление в кВт·ч
	MonthlyCost      float64 `json:"monthly_cost"`      // Стоимость эксплуатации за месяц в руб
	ImageKey         string  `json:"image_key"`         // Ключ изображения
	Description      string  `json:"description"`       // Описание устройства
	Category         string  `json:"category"`          // Категория (освещение, оборудование, etc.)
}
