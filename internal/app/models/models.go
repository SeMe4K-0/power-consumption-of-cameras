package models

// Камера видеонаблюдения
type Camera struct {
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

// Камера в заявке с дополнительными полями
type OrderCamera struct {
	Camera   Camera `json:"camera"`   // Основная информация о камере
	Quantity int    `json:"quantity"` // Количество камер
	Hours    int    `json:"hours"`    // Часы работы в сутки
	Comment  string `json:"comment"`  // Комментарий
}

// Заявка на расчет энергопотребления с вложенным массивом камер
type EnergyOrder struct {
	ID          int           `json:"id"`
	TotalPower  float64       `json:"total_power"`  // Общая мощность (Вт)
	DailyEnergy float64       `json:"daily_energy"` // Суточное потребление (кВт·ч)
	MonthlyCost float64       `json:"monthly_cost"` // Стоимость за месяц (руб)
	Tariff      float64       `json:"tariff"`       // Тариф на электроэнергию
	Status      string        `json:"status"`
	ClientName  string        `json:"client_name"`  // Имя клиента
	ProjectName string        `json:"project_name"` // Название проекта
	Cameras     []OrderCamera `json:"cameras"`      // Вложенный массив камер
}

// Расчет электроэнергии (аналог частицы)
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

// Связь заявка-услуга (отдельная модель для соблюдения 1НФ)
type OrderService struct {
	OrderID  int    `json:"order_id"`
	CameraID int    `json:"camera_id"`
	Quantity int    `json:"quantity"`
	Order    int    `json:"order"`   // Порядок в системе
	Comment  string `json:"comment"` // Комментарий
	Other    string `json:"other"`   // Другое
}

// Данные для формы заявки
type OrderFormData struct {
	Order            EnergyOrder         `json:"order"`
	AvailableCameras []Camera            `json:"available_cameras"`
	Calculations     []EnergyCalculation `json:"calculations"`
}

// Связь заявка-услуга с полной информацией о камере
type OrderServiceWithCamera struct {
	OrderService
	Camera Camera `json:"camera"`
}
