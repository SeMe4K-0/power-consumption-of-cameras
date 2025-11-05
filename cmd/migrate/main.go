package main

import (
	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загружаем переменные окружения из app.env
	_ = godotenv.Load("app.env")
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.Cameras{},
		&ds.RequestCamerasCalculation{},
		&ds.CamerasCalculation{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
