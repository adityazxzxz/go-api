package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-api/config"
	"go-api/utils"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var secretKey = "mysupersecretkeymustbe32bytes!!!"
var jwtKey = []byte(secretKey)

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
			return jwtKey, nil
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

		decryptedJSON, err := utils.Decrypt(claims.Data, jwtKey)
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
	encryptedData, err := utils.Encrypt(string(jsonBytes), jwtKey)
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
	return token.SignedString(jwtKey)
}

func HMACAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		signature := c.GetHeader("X-Signature")
		nonce := c.GetHeader("X-Nonce")
		timestamp := c.GetHeader("X-Timestamp")

		if signature == "" || nonce == "" || timestamp == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing HMAC headers",
			})
			return
		}

		args := redis.SetArgs{
			Mode: "NX",             // <- ini pengganti NX
			TTL:  60 * time.Second, // 60 detik untuk mencegah replay attack
		}

		err := config.Redis.SetArgs(
			config.Ctx,
			nonce,
			1,
			args,
		).Err()

		if err == redis.Nil {
			fmt.Println("too much request, please try again later")
			return
		}

		if err != nil {
			panic(err)
		}

		tsInt, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid timestamp",
			})
			return
		}

		now := time.Now().Unix()
		if now-tsInt > 300 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Request expired",
			})
			return
		}

		if tsInt-now > 300 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid timestamp (future)",
			})
			return
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		body := string(bodyBytes)

		method := strings.ToUpper(c.Request.Method)
		payload := method + ":" + nonce + ":" + timestamp + ":" + body

		mac := hmac.New(sha256.New, []byte(secretKey))
		mac.Write([]byte(payload))
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(expectedMAC), []byte(signature)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid signature",
			})
			return
		}

		c.Next()
	}
}
