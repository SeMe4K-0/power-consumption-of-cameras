package main

import (
	"log"

	"Lab1/internal/api" // Импортируем наш пакет api
)

func main() {
	log.Println("Application start!")
	api.StartServer() // Запускаем сервер
	log.Println("Application terminated!")
}
