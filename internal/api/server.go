package api

import (
	"it-maintenance-backend/internal/app/dsn"
	"it-maintenance-backend/internal/app/handler"
	"it-maintenance-backend/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	postgresString := dsn.FromEnv()
	repo, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	handlers := handler.NewHandler(repo)

	r := gin.Default()

	r.SetFuncMap(map[string]interface{}{
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
	})

	r.Static("/static", "./resources")
	r.LoadHTMLGlob("./templates/*")

	r.GET("/cameras", handlers.GetServices)
	r.GET("/camera/:id", handlers.GetServiceDetail)
	r.GET("/electricity-calculation/:id", handlers.GetCurrentOrder)
	r.POST("/order/add-service", handlers.AddServiceToOrder)
	r.POST("/order/delete/:id", handlers.DeleteOrder)

	log.Println("Сервер запущен на http://localhost:8080")
	log.Println("Доступные страницы:")
	log.Println("  GET /cameras - Список камер")
	log.Println("  GET /camera/:id - Детали камеры")
	log.Println("  GET /electricity-calculation/:id - Детали заявки")
	log.Println("  POST /order/add-service - Добавить услугу в заявку")
	log.Println("  POST /order/delete/:id - Удалить заявку")

	r.Run()
	log.Println("Server down")
}
