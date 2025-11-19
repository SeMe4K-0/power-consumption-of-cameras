package pkg

import (
	"fmt"

	_ "awesomeProject/docs"
	"awesomeProject/internal/app/config"
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
	Redis   *redis.Client
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler, redisClient *redis.Client) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
		Redis:   redisClient,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Handler.RegisterAPI(a.Router)

	a.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
