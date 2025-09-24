package handler

import (
	"Lab1/internal/app/models"
	"Lab1/internal/app/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// Главная страница со списком камер
func (h *Handler) GetCameras(ctx *gin.Context) {
	var cameras []models.Camera
	var err error

	searchQuery := ctx.Query("search")

	if searchQuery != "" {
		cameras, err = h.Repository.GetCamerasBySearch(searchQuery)
	} else {
		cameras, err = h.Repository.GetCameras()
	}

	if err != nil {
		logrus.Error(err)
		cameras = []models.Camera{}
	}

	ordersCount := h.Repository.GetOrdersCount()

	ctx.HTML(http.StatusOK, "cameras.html", gin.H{
		"time":        time.Now().Format("02.01.2006 15:04"),
		"cameras":     cameras,
		"searchQuery": searchQuery,
		"ordersCount": ordersCount,
	})
}

// Страница деталей камеры
func (h *Handler) GetCameraDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID камеры: ", err)
		ctx.Redirect(http.StatusFound, "/cameras")
		return
	}

	camera, err := h.Repository.GetCameraByID(id)
	if err != nil {
		logrus.Error("Камера не найдена: ", err)
		ctx.Redirect(http.StatusFound, "/cameras")
		return
	}

	ordersCount := h.Repository.GetOrdersCount()

	ctx.HTML(http.StatusOK, "camera_detail.html", gin.H{
		"camera":      camera,
		"time":        time.Now().Format("02.01.2006 15:04"),
		"ordersCount": ordersCount,
	})
}

// Страница заявки на расчет
func (h *Handler) GetOrderDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID заявки: ", err)
		ctx.Redirect(http.StatusFound, "/cameras")
		return
	}

	formData, err := h.Repository.GetOrderFormData(id)
	if err != nil {
		logrus.Error("Ошибка получения данных заявки: ", err)
		ctx.Redirect(http.StatusFound, "/cameras")
		return
	}

	ordersCount := h.Repository.GetOrdersCount()

	ctx.HTML(http.StatusOK, "order_detail.html", gin.H{
		"order":            formData.Order,
		"availableCameras": formData.AvailableCameras,
		"orderServices":    formData.OrderServices,
		"time":             time.Now().Format("01.01.2003 15:04"),
		"ordersCount":      ordersCount,
	})
}
