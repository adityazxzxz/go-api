package main

import (
	"go-api/config"
	"go-api/controllers"
	"go-api/helpers"
	"go-api/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()         // load environment variables & key secret
	config.InitRedis()       // initial redis connection
	db, _ := config.DBInit() // initial database connection
	gin.SetMode(os.Getenv("GIN_MODE"))
	helpers.SetupLogging()

	controllers := &controllers.InDB{DB: db}

	router := gin.Default()
	router.POST("/login", controllers.Login)
	router.POST("/refresh", controllers.Refresh)

	protected := router.Group("/")
	protected.Use(
		middleware.JWTAuth(),
		middleware.HMACAuth())
	{
		protected.POST("/logout", controllers.RevokeToken)
		protected.GET("/profile", controllers.Profile)
		protected.PUT("/profile", controllers.UpdateProfile)
	}
	router.GET("/panic-test", func(c *gin.Context) {
		panic("this is test panic error log")
	})

	limiterTest := router.Group("/")
	limiterTest.Use(middleware.RateLimitByIP())

	limiterTest.GET("/limit", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "This is a rate-limited endpoint",
		})
	})

	router.Run(":" + os.Getenv("APP_PORT"))
}
