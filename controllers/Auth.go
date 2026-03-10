package controllers

import (
	"errors"
	"fmt"
	"go-api/middleware"
	"go-api/models"
	"go-api/resources"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	var session models.UserSessions
	expiredTimeStr := os.Getenv("JWT_EXPIRE")
	expiredTime, err := strconv.Atoi(expiredTimeStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var result Result

	err = idb.DB.
		Table("user_sessions").
		Select("user_sessions.user_id, users.email").
		Joins("JOIN users ON users.id = user_sessions.user_id").
		Where("user_sessions.refresh_token = ? AND user_sessions.revoked = ? AND user_sessions.expired_at > ?", req.RefreshToken, 0, time.Now().Unix()).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Session not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	refresh_token, err := middleware.GenerateRefreshToken(idb.DB, result.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	resultSession := idb.DB.Model(&session).Where("refresh_token = ?", req.RefreshToken).Updates(map[string]interface{}{
		"refresh_token": refresh_token,
	})
	if resultSession.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update refresh token"})
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
		ExpiresIn:    expiredTime * 3600,
	}

	c.JSON(http.StatusOK, response)
}

func (idb *InDB) Login(c *gin.Context) {
	var req LoginRequest
	expiredTimeStr := os.Getenv("JWT_EXPIRE")
	refreshTokenExpireStr := os.Getenv("REFRESH_TOKEN_EXPIRE_DAYS")
	expiredTime, err := strconv.Atoi(expiredTimeStr)
	refreshTokenExpire, err := strconv.Atoi(refreshTokenExpireStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

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

	token, err := middleware.GenerateJWT(user.ID, user.Email, expiredTime)
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
		UserAgent:    c.Request.UserAgent(),
		ExpiredAt:    time.Now().Add(time.Duration(refreshTokenExpire) * 24 * time.Hour).Unix(),
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
		ExpiresIn:    expiredTime * 3600, // dalam detik
	}

	c.JSON(http.StatusOK, response)
}

func (idb *InDB) RevokeToken(c *gin.Context) {
	var req RefreshRequest
	var session models.UserSessions

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// err := idb.DB.Where("user_id = ?", userID).Delete(&models.UserSessions{}).Error
	resultSession := idb.DB.Model(&session).Where("refresh_token = ? AND user_id = ?", req.RefreshToken, userID).Updates(map[string]interface{}{
		"revoked": 1,
	})

	if resultSession.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
}
