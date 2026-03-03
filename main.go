package main

import (
	"go-api/config"
	"go-api/controllers"
	"go-api/middleware"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	config.LoadEnv()         // load environment variables & key secret
	config.InitRedis()       // initial redis connection
	db, _ := config.DBInit() // initial database connection
	gin.SetMode(os.Getenv("GIN_MODE"))
	setupLogging()

	controllers := &controllers.InDB{DB: db}

	router := gin.Default()
	router.POST("/login", controllers.Login)

	protected := router.Group("/")
	protected.Use(middleware.JWTAuth(), middleware.HMACAuth())
	{
		protected.GET("/profile", controllers.Profile)
		protected.PUT("/profile", controllers.UpdateProfile)
	}
	router.GET("/panic-test", func(c *gin.Context) {
		panic("this is test panic error log")
	})

	router.Run(":" + os.Getenv("APP_PORT"))
}

func setupLogging() {

	// Pastikan folder logs ada
	os.MkdirAll("logs", os.ModePerm)

	// Access log file
	accessLog := &lumberjack.Logger{
		Filename:   "logs/access.log",
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // hari
		Compress:   true,
	}

	// Error log file
	errorLog := &lumberjack.Logger{
		Filename:   "logs/error.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	// Set output
	gin.DefaultWriter = io.MultiWriter(accessLog, os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(errorLog, os.Stderr)
}
