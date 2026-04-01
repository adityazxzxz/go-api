package controllers

import (
	"context"
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
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
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
		helpers.ErrorLogger.Println("Refresh token error:", err)
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
		helpers.ErrorLogger.Println("Invalid token expiration configuration:", err)
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
		var otpEmailBody string

		otpResponse, otpCode, err := idb.createOTP(c, &user, otpEmailBody)
		if err != nil {
			helpers.ErrorLogger.Println("OTP generation error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate otp"})
			return
		}

		err = idb.DB.
			Model(&models.EmailTemplate{}).
			Select("body").
			Where("template_name = ?", "otp").
			Scan(&otpEmailBody).Error

		helpers.ErrorLogger.Println("error email template", err)

		if err == nil {
			mailPayload := helpers.MailTemplateFormat(map[string]interface{}{
				"nama": req.Email,
				"kode": otpCode,
			}, otpEmailBody)
			go helpers.SendEmail(req.Email, "Subject", mailPayload)
		}

		response = otpResponse
	} else {
		tokenResponse, err := createToken(c, idb, &user, expiredTime, refreshTokenExpire)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		response = tokenResponse
	}

	c.JSON(http.StatusOK, response)
}

func (idb *InDB) LoginMagicLinkRequest(c *gin.Context) {
	var response resources.MagicLinkResponse
	var req requests.MagicLinkRequest
	var body string

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	token, err := SendMagicLink(req.Email)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send magic link"})
		return
	}

	err = idb.DB.
		Model(&models.EmailTemplate{}).
		Select("body").
		Where("template_name = ?", "magic_link").
		Scan(&body).Error

	if err == nil {
		mailPayload := helpers.MailTemplateFormat(map[string]interface{}{
			"nama": req.Email,
			"kode": token,
		}, body)
		go helpers.SendEmail(req.Email, "Subject", mailPayload)
	}

	// region send email token
	// endregion

	response = resources.MagicLinkResponse{
		Error:   false,
		Message: "Magic link sent successfully",
		Data: resources.MagicLinkData{
			MagicToken: token,
			URL:        os.Getenv("FRONTEND_URL") + "/verify-link?token=" + token,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (idb *InDB) VerifyMagicLink(c *gin.Context) {
	var req requests.VerifyMagicLinkRequest
	var tokenResponse any

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	token := req.MagicToken
	ctx := context.Background()
	key := "magic_link:" + token

	storedData, err := config.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Magic link expired or not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error"})
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(storedData), &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse magic link data"})
		return
	}

	email := data["email"].(string)
	fmt.Println("email dari redis", email)

	var user models.User
	err = idb.DB.Where("email = ?", email).First(&user).Error
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				Email:     email,
				LoginType: "magic_link",
				UUID:      uuid.NewString(),
				Status:    1,
				LastLogin: time.Now().Unix(),
				LastIP:    c.ClientIP(),
			}

			if err := idb.DB.Create(&user).Error; err != nil {
				fmt.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
				return
			}

		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
	}

	updates := models.User{
		LastLogin: time.Now().Unix(),
		LastIP:    c.ClientIP(),
	}

	err = idb.DB.Model(&user).Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.Response{
			Error:   true,
			Message: "Update last login failed",
			Data:    nil,
		})

		return
	}

	tokenResponse, err = createToken(c, idb, &user, 1, 7)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Hapus magic link setelah dipakai (single use)
	config.Redis.Del(ctx, key)

	c.JSON(http.StatusOK, tokenResponse)
}

func SendMagicLink(email string) (string, error) {
	token, err := helpers.GenerateMagicToken()

	data := map[string]interface{}{
		"email": email,
	}
	ctx := context.Background()
	jsonData, _ := json.Marshal(data)

	// Simpan magic link di Redis dengan TTL 5 menit
	err = config.Redis.Set(
		ctx,
		"magic_link:"+token,
		jsonData,
		5*time.Minute,
	).Err()

	// err = config.Redis.Set(c.Request.Context(), "otp:"+challengeID, otpCode, 5*time.Minute).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (idb *InDB) LoginGoogle(c *gin.Context) {
	var req requests.GoogleLoginRequest
	var response any
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	payload, err := idtoken.Validate(
		context.Background(),
		req.IDToken,
		googleClientID,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid google token",
		})
		return
	}
	email := payload.Claims["email"].(string)
	err = idb.DB.Where("email = ?", email).First(&models.User{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user := models.User{
				Email:     email,
				FirstName: payload.Claims["name"].(string),
				LoginType: "google",
				UUID:      uuid.NewString(),
				Status:    1,
				LastLogin: time.Now().Unix(),
				LastIP:    c.ClientIP(),
			}

			if err := idb.DB.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to create user",
				})
				return
			}
		}
		name := payload.Claims["name"].(string)
		// googleID := payload.Subject
		var user models.User
		user = models.User{
			Email:     email,
			FirstName: name,
			LoginType: "google",
			UUID:      uuid.NewString(),
			Status:    1,
			LastLogin: time.Now().Unix(),
			LastIP:    c.ClientIP(),
		}

		tokenResponse, err := createToken(c, idb, &user, 1, 7)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		response = tokenResponse

		c.JSON(http.StatusOK, response)
	}
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
	var response any
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
	user := models.User{
		ID:    userID,
		Email: email,
	}
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

	tokenResponse, err := createToken(c, idb, &user, 1, 7)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	response = tokenResponse

	// Hapus OTP setelah dipakai (single use)
	config.Redis.Del(ctx, key)

	c.JSON(http.StatusOK, response)

}

func createToken(c *gin.Context, idb *InDB, user *models.User, expiredTime int, refreshTokenExpire int) (any, error) {
	token, err := middleware.GenerateJWT(user.ID, user.Email, expiredTime)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return "", err
	}

	refresh_token, err := middleware.GenerateRefreshToken(idb.DB, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return "", err
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
		return "", err
	}
	tokenResponse := resources.ResponseLogin{
		Error:        false,
		Message:      "Login successful",
		AccessToken:  token,
		RefreshToken: refresh_token,
		ExpiresIn:    expiredTime * 3600, // dalam detik
	}
	return tokenResponse, nil
}

func (idb *InDB) createOTP(c *gin.Context, user *models.User, emailBody string) (any, string, error) {
	ttl, err := strconv.Atoi(os.Getenv("OTP_TTL_MINUTES"))
	challengeID, otpCode := helpers.GenerateOTP()
	data := map[string]interface{}{
		"otp":     otpCode,
		"user_id": user.ID,
		"email":   user.Email,
	}

	jsonData, _ := json.Marshal(data)

	// Simpan OTP di Redis dengan TTL set di env
	err = config.Redis.Set(
		c.Request.Context(),
		"otp:"+challengeID,
		jsonData,
		time.Duration(ttl)*time.Minute,
	).Err()

	// err = config.Redis.Set(c.Request.Context(), "otp:"+challengeID, otpCode, 5*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP"})
		return nil, "", err
	}

	otpResponse := resources.OtpResponse{
		Error:       false,
		Message:     "OTP sent successfully",
		ChallengeID: challengeID,
	}

	if os.Getenv("GIN_MODE") != "release" {
		otpResponse.OtpDebug = otpCode
	}

	return otpResponse, otpCode, nil
}

func (idb *InDB) PayloadTest(c *gin.Context) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Gagal membaca body",
		})
		return
	}

	// Restore body (biar bisa dipakai lagi)

	// Parse ke JSON (map)
	var jsonData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Body bukan JSON valid",
		})
		return
	}

	// Return sebagai JSON
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "HMAC valid, body diterima",
		"data":    jsonData,
	})
}
