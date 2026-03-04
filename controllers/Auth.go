package controllers

import (
	"fmt"
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
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type Result struct {
	UserID uint
	Email  string
}

func (idb *InDB) Refresh(c *gin.Context) {
	var req RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var result Result

	err := idb.DB.
		Table("user_sessions").
		Select("user_sessions.user_id, users.email").
		Joins("JOIN users ON users.id = user_sessions.user_id").
		Where("user_sessions.refresh_token = ?", req.RefreshToken).
		Scan(&result).Error

	refresh_token, err := middleware.GenerateRefreshToken(idb.DB, result.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	access_token, err := middleware.GenerateJWT(result.UserID, result.Email, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	response := resources.ResponseRefresh{
		Error:        false,
		Message:      "Token refreshed successfully",
		AccessToken:  access_token,
		RefreshToken: refresh_token,
	}

	c.JSON(http.StatusOK, response)
}

func (idb *InDB) Login(c *gin.Context) {
	var req LoginRequest
	expireTime := 1 // dalam jam

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

	token, err := middleware.GenerateJWT(user.ID, user.Email, expireTime)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refresh_token, err := middleware.GenerateRefreshToken(idb.DB, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	refresh := models.UserSessions{
		UserID:       user.ID,
		RefreshToken: refresh_token,
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	err = idb.DB.Create(&refresh).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	idb.DB.Model(&user).Updates(map[string]interface{}{
		"last_login": time.Now().Unix(),
		"last_ip":    c.ClientIP(),
	})

	response := resources.ResponseLogin{
		Error:        false,
		Message:      "Login successful",
		AccessToken:  token,
		RefreshToken: refresh_token,
		ExpiresIn:    expireTime * 3600, // dalam detik
	}

	c.JSON(http.StatusOK, response)
}
