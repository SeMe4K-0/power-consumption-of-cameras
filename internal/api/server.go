package api

import (
	"Lab1/internal/app/handler"
	"Lab1/internal/app/repository"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория: ", err)
	}

	handlers := handler.NewHandler(repo)

	r := gin.Default()

	// Добавляем функции для шаблонов
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
		"imageURL": func(key string) string {
			if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
				return key
			}
			return "/static/img/cameras/" + key
		},
	})

	// Статические файлы
	r.Static("/static", "./resources")
	r.LoadHTMLGlob("./templates/*")

	// Три Get запроса
	r.GET("/cameras", handlers.GetCameras)
	r.GET("/camera/:id", handlers.GetCameraDetail)
	r.GET("/electricity-calculation/:id", handlers.GetOrderDetail)

	log.Println("Сервер запущен на http://localhost:8080")
	log.Println("Доступные страницы:")
	log.Println("  GET /cameras - Список камер")
	log.Println("  GET /camera/:id - Детали камеры")
	log.Println("  GET /electricity-calculation/:id - Детали заявки")

	r.Run()
	log.Println("Server down")
}
