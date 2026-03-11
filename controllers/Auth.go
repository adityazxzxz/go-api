package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-api/config"
	"go-api/helpers"
	"go-api/middleware"
	"go-api/models"
	"go-api/requests"
	"go-api/resources"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Result struct {
	UserID uint
	Email  string
}

func (idb *InDB) Refresh(c *gin.Context) {
	var req requests.RefreshRequest
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
	var response any
	var req requests.LoginRequest
	otp := os.Getenv("OTP")
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

	idb.DB.Model(&user).Updates(map[string]interface{}{
		"last_login": time.Now().Unix(),
		"last_ip":    c.ClientIP(),
	})

	// region generate token
	if otp == "1" {
		challengeID, otpCode := helpers.GenerateOTP()
		data := map[string]interface{}{
			"otp":     otpCode,
			"user_id": user.ID,
			"email":   user.Email,
		}

		jsonData, _ := json.Marshal(data)

		// Simpan OTP di Redis dengan TTL 5 menit
		err = config.Redis.Set(
			c.Request.Context(),
			"otp:"+challengeID,
			jsonData,
			5*time.Minute,
		).Err()

		// err = config.Redis.Set(c.Request.Context(), "otp:"+challengeID, otpCode, 5*time.Minute).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP"})
			return
		}

		otpResponse := resources.OtpResponse{
			Error:       false,
			Message:     "OTP sent successfully",
			ChallengeID: challengeID,
		}

		if os.Getenv("GIN_MODE") != "release" {
			otpResponse.OtpDebug = otpCode
		}
		response = otpResponse
	} else {
		token, refresh_token, err := createToken(c, idb, user.ID, user.Email, expiredTime, refreshTokenExpire)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		tokenResponse := resources.ResponseLogin{
			Error:        false,
			Message:      "Login successful",
			AccessToken:  token,
			RefreshToken: refresh_token,
			ExpiresIn:    expiredTime * 3600, // dalam detik
		}
		response = tokenResponse
	}

	c.JSON(http.StatusOK, response)
}

func (idb *InDB) RevokeToken(c *gin.Context) {
	var req requests.RefreshRequest
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

func (idb *InDB) VerifyOTP(c *gin.Context) {
	var req requests.VerifyOtpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// Ambil OTP dari Redis
	ctx := c.Request.Context()

	key := "otp:" + req.ChallengeID

	storedData, err := config.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired or not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error"})
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(storedData), &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse OTP data"})
		return
	}

	userID := uint(data["user_id"].(float64))
	email := data["email"].(string)
	storedOtp, ok := data["otp"].(string)

	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid OTP format"})
		return
	}

	// Compare OTP
	if storedOtp != req.Otp {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	token, refresh_token, err := createToken(c, idb, userID, email, 1, 7)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Hapus OTP setelah dipakai (single use)
	config.Redis.Del(ctx, key)

	response := resources.ResponseLogin{
		Error:        false,
		Message:      "Login successful",
		AccessToken:  token,
		RefreshToken: refresh_token,
		ExpiresIn:    3600, // dalam detik
	}

	c.JSON(http.StatusOK, response)

}

func createToken(c *gin.Context, idb *InDB, userID uint, email string, expiredTime int, refreshTokenExpire int) (string, string, error) {
	token, err := middleware.GenerateJWT(userID, email, expiredTime)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return "", "", err
	}

	refresh_token, err := middleware.GenerateRefreshToken(idb.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return "", "", err
	}

	refresh := models.UserSessions{
		UserID:       userID,
		RefreshToken: refresh_token,
		UserAgent:    c.Request.UserAgent(),
		ExpiredAt:    time.Now().Add(time.Duration(refreshTokenExpire) * 24 * time.Hour).Unix(),
	}

	err = idb.DB.Create(&refresh).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return "", "", err
	}
	return token, refresh_token, nil
}

func createOTP() any {
	challengeID, otpCode := helpers.GenerateOTP()
	response := resources.OtpResponse{
		Error:       false,
		Message:     "Req OTP successful",
		ChallengeID: challengeID,
	}
	if os.Getenv("APP_ENV") != "production" {
		response.OtpDebug = otpCode
	}

	return response
}
