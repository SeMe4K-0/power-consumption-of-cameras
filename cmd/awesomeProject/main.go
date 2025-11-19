package main

import (
	"awesomeProject/internal/app/config"
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/redis"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/pkg"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title Cameras Calculation API
// @version 1.0
// @description API для расчета камер видеонаблюдения

// @host 127.0.0.1:8080
// @schemes http https
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MB

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()

	rep, errRep := repository.New(
		postgresString,
		conf.MinIO.Endpoint,
		conf.MinIO.AccessKeyID,
		conf.MinIO.SecretAccessKey,
		conf.MinIO.BucketName,
		conf.MinIO.UseSSL,
	)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	ctx := context.Background()
	redisClient, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	hand := handler.NewHandler(rep, conf, func(requireModerator bool) gin.HandlerFunc {
		tempApp := pkg.NewApp(conf, router, nil, redisClient)
		return tempApp.WithAuthCheck(requireModerator)
	}, func() gin.HandlerFunc {
		tempApp := pkg.NewApp(conf, router, nil, redisClient)
		return tempApp.WithOptionalAuthCheck()
	}, redisClient)

	application := pkg.NewApp(conf, router, hand, redisClient)
	application.RunApp()
}
