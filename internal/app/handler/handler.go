package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"it-maintenance-backend/internal/app/models"
	"it-maintenance-backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.Static("/static", "./resources")

	router.SetFuncMap(map[string]interface{}{
		"multiply": func(a, b float64) float64 {
			return a * b
		},
		"divide": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"float64": func(i int) float64 {
			return float64(i)
		},
		"imageURL": func(key *string) string {
			if key == nil {
				return "/static/img/placeholder.jpg"
			}
			if strings.HasPrefix(*key, "http://") || strings.HasPrefix(*key, "https://") {
				return *key
			}
			return "/static/img/cameras/" + *key
		},
		"translateType": func(typeStr string) string {
			switch typeStr {
			case "Indoor":
				return "Внутренняя"
			case "Outdoor":
				return "Уличная"
			case "Equipment":
				return "Оборудование"
			default:
				return typeStr
			}
		},
		"translateDescription": func(name, desc string) string {
			descriptions := map[string]string{
				"Dahua IPC-HFW2831T-ZS":         "8-мегапиксельная уличная купольная камера с зумом. Защита IP67, работа при -40°C до +60°C.",
				"Hikvision DS-2CD2043G0-I":      "4-мегапиксельная купольная камера для помещений. Инфракрасная подсветка до 30м, вандалоустойчивый корпус.",
				"Axis M3045-V":                  "2-мегапиксельная купольная камера для помещений. Компактный дизайн, простое управление.",
				"TP-Link Tapo C310":             "3-мегапиксельная поворотная камера для улицы. Wi-Fi подключение, мобильное приложение.",
				"Reolink RLC-811A":              "8-мегапиксельная уличная камера с ИК подсветкой. Защита IP66, детекция движения.",
				"Dahua NVR4108-8P-4KS2":         "8-канальный видеорегистратор с поддержкой 4K. Встроенный PoE коммутатор, удаленный доступ.",
				"Hikvision DS-7608NI-K2/8P":     "8-канальный NVR с поддержкой 8MP. Встроенный PoE, детекция движения, мобильное приложение.",
				"Ubiquiti UniFi Protect G4 Pro": "Профессиональная IP камера 4K с ИК подсветкой. Интеграция с UniFi Protect, детекция движения.",
			}
			if correctDesc, ok := descriptions[name]; ok {
				return correctDesc
			}
			return desc
		},
		"cartClass": func(count int) string {
			if count > 0 {
				return "has-items"
			}
			return "empty"
		},
	})

	router.LoadHTMLGlob("./templates/*")

	router.GET("/cameras", h.GetServices)
	router.GET("/camera/:id", h.GetServiceDetail)
	router.GET("/electricity-calculation/:id", h.GetCurrentOrder)
	router.POST("/order/add-service", h.AddServiceToOrder)
	router.POST("/order/delete/:id", h.DeleteOrder)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources") // Изменено с /styles на /static и ./styles на ./resources
}

func (h *Handler) GetServices(ctx *gin.Context) {
	userID := uint(2) // Используем user1, у которого есть заявки

	var cameras []models.Camera
	var err error

	searchQuery := ctx.Query("camerasearch")

	if searchQuery != "" {
		cameras, err = h.Repository.GetCamerasBySearch(searchQuery)
	} else {
		cameras, err = h.Repository.GetCameras()
	}

	if err != nil {
		logrus.Error(err)
		cameras = []models.Camera{}
	}

	logrus.Infof("Loaded %d cameras", len(cameras))

	servicesCount := h.Repository.GetCurrentOrderServicesCount(userID)
	firstOrderID := h.Repository.GetFirstOrderID(userID)
	hasDraftOrder := h.Repository.HasDraftOrder(userID)

	hasActiveOrder := hasDraftOrder && servicesCount > 0

	ctx.HTML(http.StatusOK, "cameras.html", gin.H{
		"time":           time.Now().Format("02.01.2006 15:04"),
		"cameras":        cameras,
		"searchQuery":    searchQuery,
		"servicesCount":  servicesCount,
		"firstOrderID":   firstOrderID,
		"hasDraftOrder":  hasDraftOrder,
		"hasActiveOrder": hasActiveOrder,
		"userID":         userID,
	})
}

func (h *Handler) GetServiceDetail(ctx *gin.Context) {
	userID := uint(2) // Используем user1, у которого есть заявки

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID камеры: ", err)
		ctx.Redirect(http.StatusFound, "/cameras")
		return
	}

	camera, err := h.Repository.GetCameraByID(uint(id))
	if err != nil {
		logrus.Error("Камера не найдена: ", err)
		ctx.Redirect(http.StatusFound, "/cameras")
		return
	}

	servicesCount := h.Repository.GetCurrentOrderServicesCount(userID)
	firstOrderID := h.Repository.GetFirstOrderID(userID)
	hasDraftOrder := h.Repository.HasDraftOrder(userID)

	ctx.HTML(http.StatusOK, "camera_detail.html", gin.H{
		"service":       camera, // Используем старое название для совместимости с шаблонами
		"time":          time.Now().Format("02.01.2006 15:04"),
		"servicesCount": servicesCount,
		"firstOrderID":  firstOrderID,
		"hasDraftOrder": hasDraftOrder,
		"userID":        userID,
	})
}

func (h *Handler) GetCurrentOrder(ctx *gin.Context) {
	userID := uint(2) // Используем user1, у которого есть заявки

	idStr := ctx.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logrus.Error("Неверный ID заявки: ", err)
		ctx.Redirect(http.StatusFound, "/cameras")
		return
	}

	err = h.Repository.CheckOrderAccess(uint(orderID), userID)
	if err != nil {
		logrus.Warnf("Доступ к заявке %d запрещен для пользователя %d: %v", orderID, userID, err)
		ctx.HTML(http.StatusForbidden, "error.html", gin.H{
			"error":   "Доступ к заявке запрещен",
			"message": "Заявка не найдена, не принадлежит вам или пуста",
		})
		return
	}

	order, err := h.Repository.GetOrderByID(uint(orderID))
	if err != nil {
		logrus.Error("Заявка не найдена: ", err)
		ctx.Redirect(http.StatusFound, "/cameras")
		return
	}

	orderCameras, err := h.Repository.GetOrderCameras(uint(orderID))
	if err != nil {
		logrus.Error("Ошибка получения камер заявки: ", err)
		orderCameras = []models.OrderCamera{}
	}

	logrus.Infof("Loaded order %d with %d cameras", orderID, len(orderCameras))

	servicesCount := h.Repository.GetOrderServicesCount(uint(orderID))
	firstOrderID := h.Repository.GetFirstOrderID(userID)

	ctx.HTML(http.StatusOK, "order_detail.html", gin.H{
		"order":         order,
		"orderCameras":  orderCameras,
		"time":          time.Now().Format("02.01.2006 15:04"),
		"servicesCount": servicesCount,
		"firstOrderID":  firstOrderID,
		"userID":        userID,
	})
}

func (h *Handler) AddServiceToOrder(ctx *gin.Context) {
	userID := uint(2) // Используем user1, у которого есть заявки

	var request struct {
		ServiceID   uint   `json:"service_id" binding:"required"`
		Quantity    int    `json:"quantity" binding:"required,min=1"`
		Comment     string `json:"comment"`
		Other       string `json:"other"`
		ClientName  string `json:"client_name"`
		ProjectName string `json:"project_name"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.SurveillanceOrder
	var err error
	if h.Repository.HasDraftOrder(userID) {
		order, err = h.Repository.GetCurrentOrder(userID)
		if err != nil {
			logrus.Error("Ошибка получения текущей заявки: ", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения заявки"})
			return
		}
	} else {
		order, err = h.Repository.CreateOrder(userID, request.ClientName, request.ProjectName)
		if err != nil {
			logrus.Error("Ошибка создания заявки: ", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания заявки"})
			return
		}
	}

	err = h.Repository.AddCameraToOrder(order.ID, request.ServiceID, request.Quantity, request.Comment, request.Other)
	if err != nil {
		logrus.Error("Ошибка добавления камеры в заявку: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления камеры"})
		return
	}

	servicesCount := h.Repository.GetCurrentOrderServicesCount(userID)

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Камера успешно добавлена в заявку",
		"servicesCount": servicesCount,
	})
}

func (h *Handler) DeleteOrder(ctx *gin.Context) {
	userID := uint(2) // Используем user1, у которого есть заявки

	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID заявки: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заявки"})
		return
	}

	order, err := h.Repository.GetOrderByID(uint(orderID))
	if err != nil {
		logrus.Error("Заявка не найдена: ", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
		return
	}

	if order.CreatorID != userID || order.Status != models.OrderStatusDraft {
		logrus.Warnf("Пользователь %d пытается удалить чужую заявку или заявку не в статусе черновик: %d", userID, orderID)
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав для удаления заявки"})
		return
	}

	err = h.Repository.DeleteOrder(uint(orderID))
	if err != nil {
		logrus.Error("Ошибка удаления заявки: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления заявки"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка успешно удалена"})
}
