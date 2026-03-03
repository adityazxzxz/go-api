package main

import (
	"go-api/config"
	"go-api/controllers"
	"go-api/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()         // load environment variables & key secret
	config.InitRedis()       // initial redis connection
	db, _ := config.DBInit() // initial database connection

	controllers := &controllers.InDB{DB: db}

	router := gin.Default()
	router.POST("/login", controllers.Login)

	protected := router.Group("/")
	protected.Use(middleware.JWTAuth(), middleware.HMACAuth())
	{
		protected.GET("/profile", controllers.Profile)
		protected.PUT("/profile", controllers.UpdateProfile)
	}

	router.Run(":" + os.Getenv("APP_PORT"))
}
