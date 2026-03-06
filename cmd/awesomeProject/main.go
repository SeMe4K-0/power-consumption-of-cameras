package main

import (
	"fmt"
	"log"

	"it-maintenance-backend/internal/app/config"
	"it-maintenance-backend/internal/app/dsn"
	"it-maintenance-backend/internal/app/handler"
	"it-maintenance-backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	log.Println("Application start!")

	_ = godotenv.Load("config.env")

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println("Connecting to database...")
	fmt.Println(postgresString)

	repo, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	handlers := handler.NewHandler(repo)

	router := gin.Default()
	handlers.RegisterHandler(router)

	serverAddress := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)
	log.Printf("Сервер запущен на http://%s", serverAddress)
	log.Println("Доступные страницы:")
	log.Println("  GET /services - Список услуг")
	log.Println("  GET /service/:id - Детали услуги")
	log.Println("  GET /order - Текущая заявка")
	log.Println("  POST /order/add-service - Добавить услугу в заявку")
	log.Println("  POST /order/delete/:id - Удалить заявку")

	if err := router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}

	log.Println("Application terminated!")
}
