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

func main() {
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	router.MaxMultipartMemory = 8 << 20

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
