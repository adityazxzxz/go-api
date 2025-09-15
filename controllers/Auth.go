package controllers

import (
	"go-api/middleware"
	"go-api/models"
	"go-api/resources"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (idb *InDB) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := idb.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	expireTime := 24

	token, err := middleware.GenerateJWT(user.ID, user.Email, expireTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	idb.DB.Model(&user).Updates(map[string]interface{}{
		"last_login": time.Now().Unix(),
		"last_ip":    c.ClientIP(),
	})

	response := resources.ResponseLogin{
		Error:       false,
		Message:     "Login successful",
		AccessToken: token,
		ExpiresIn:   expireTime * 3600, // dalam detik
	}

	c.JSON(http.StatusOK, response)
}
