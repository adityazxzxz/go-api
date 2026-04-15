package main

import (
	"go-api/config"
	"go-api/controllers"
	"go-api/helpers"
	"go-api/library/i18n"
	"go-api/middleware"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	helpers.SetupLogging()
	i18n.Init()
	config.LoadEnv()         // load environment variables & key secret
	config.InitRedis()       // initial redis connection
	db, _ := config.DBInit() // initial database connection
	gin.SetMode(os.Getenv("GIN_MODE"))

	controllers := &controllers.InDB{DB: db}
	router := gin.Default()
	router.Use(middleware.LanguageMiddleware())
	router.Use(cors.Default())
	// perlu config cors untuk production, sekarang pakai default untuk development

	router.POST("/login-google", controllers.LoginGoogle)

	hmacProtect := router.Group("/")
	hmacProtect.Use(
		middleware.HMACAuth(),
		middleware.RateLimitByIP(),
	)
	{

		hmacProtect.POST("/login", controllers.Login)
		hmacProtect.POST("/login-magic-link", controllers.LoginMagicLinkRequest)
		hmacProtect.POST("/verify-link", controllers.VerifyMagicLink)
		hmacProtect.POST("/register", controllers.Register)
		hmacProtect.POST("/verify-otp", controllers.VerifyOTP)
		hmacProtect.POST("/payload-test", controllers.PayloadTest)
	}

	allProtect := router.Group("/")
	allProtect.Use(
		middleware.JWTAuth(),
		middleware.HMACAuth(),
		middleware.RateLimitByIP(),
	)
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

	// region consumer terpisah untuk miksoservice, bisa dipindah ke file lain
	// go consumers.MailConsumer()
	// endregion
	router.Run(":" + os.Getenv("APP_PORT"))
}
