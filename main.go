package main

import (
	"go-api/config"
	"go-api/controllers"
	"go-api/middleware"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	env := godotenv.Load()
	if env != nil {
		panic("Error loading .env file")
	}

	db, _ := config.DBInit()
	controllers := &controllers.InDB{DB: db}

	router := gin.Default()
	router.POST("/login", controllers.Login)

	protected := router.Group("/")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/profile", controllers.Profile)
		protected.PUT("/profile", controllers.UpdateProfile)
	}

	router.Run(":" + os.Getenv("APP_PORT"))
}
