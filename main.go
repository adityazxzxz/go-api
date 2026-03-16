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

	hmacProtect := router.Group("/")
	hmacProtect.Use(middleware.HMACAuth())
	{
		hmacProtect.POST("/login", controllers.Login)
		hmacProtect.POST("/login-magic-link", controllers.LoginMagicLinkRequest)
		hmacProtect.POST("/verify-link", controllers.VerifyMagicLink)
		hmacProtect.POST("/register", controllers.Register)
		hmacProtect.POST("/verify-otp", controllers.VerifyOTP)
	}

	allProtect := router.Group("/")
	allProtect.Use(
		middleware.JWTAuth(),
		middleware.HMACAuth())
	{
		allProtect.POST("/logout", controllers.RevokeToken)

		allProtect.POST("/refresh", controllers.Refresh)
		allProtect.GET("/profile", controllers.Profile)
		allProtect.PUT("/profile", controllers.UpdateProfile)
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
