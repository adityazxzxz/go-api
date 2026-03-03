package middleware

import (
	"encoding/json"
	"fmt"
	"go-api/config"
	"go-api/utils"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Declare key ada di config/Key.go

type UserClaims struct {
	UserID int    `json:"userid"`
	Email  string `json:"email"`
	jwt.StandardClaims
}
type Claims struct {
	Data string `json:"data"`
	jwt.StandardClaims
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "Authorization header is missing",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}
		tokenString := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.JWTKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		expirationTime := int64(claims.ExpiresAt)
		if time.Now().Unix() > expirationTime {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			c.Abort()
			return
		}
		decryptedJSON, err := utils.Decrypt(claims.Data, config.JWTKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token Payload Invalid"})
			c.Abort()
			return
		}
		var user UserClaims
		err = json.Unmarshal([]byte(decryptedJSON), &user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to parse user claims"})
			c.Abort()
			return
		}
		UserID := user.UserID
		Email := user.Email
		c.Set("user_id", UserID)
		c.Set("email", Email)

		c.Next()

	}

}

func GenerateJWT(userID int, email string, expireTime int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expireTime) * time.Hour) // berlaku 24 jam

	payload := UserClaims{
		UserID: userID,
		Email:  email,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encryptedData, err := utils.Encrypt(string(jsonBytes), config.JWTKey)
	if err != nil {
		return "", err
	}
	claims := &Claims{
		Data: encryptedData,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "my-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTKey)
}
