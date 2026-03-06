package main

import (
	"it-maintenance-backend/internal/app/dsn"
	"it-maintenance-backend/internal/app/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load("config.env")

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.Exec("DROP TABLE IF EXISTS order_cameras CASCADE")
	db.Exec("DROP TABLE IF EXISTS surveillance_orders CASCADE")
	db.Exec("DROP TABLE IF EXISTS cameras CASCADE")
	db.Exec("DROP TABLE IF EXISTS it_users CASCADE")

	err = db.AutoMigrate(
		&models.User{},
		&models.Camera{},
		&models.SurveillanceOrder{},
		&models.OrderCamera{},
	)
	if err != nil {
		panic("cant migrate db")
	}

	db.Exec("ALTER TABLE order_cameras ADD CONSTRAINT unique_order_camera UNIQUE (order_id, camera_id)")

	println("Миграция завершена успешно!")
}
