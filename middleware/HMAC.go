package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-api/config"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Declare key ada di config/Key.go

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

		// region mutex
		lockKey := "nonce:" + nonce

		err = config.Redis.SetArgs(
			config.Ctx,
			lockKey,
			1,
			redis.SetArgs{
				Mode: "NX",            // hanya set jika belum ada
				TTL:  5 * time.Minute, // sama dengan window timestamp
			},
		).Err()

		if err == redis.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Replay attack detected",
			})
			return
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Redis error",
			})
			return
		}
		// end region mutex

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		body := string(bodyBytes)

		method := strings.ToUpper(c.Request.Method)
		payload := method + ":" + nonce + ":" + timestamp + ":" + body

		mac := hmac.New(sha256.New, []byte(config.HMACSecret))
		mac.Write([]byte(payload))
		expectedMAC := mac.Sum(nil)

		sigBytes, err := hex.DecodeString(signature)
		if err != nil {

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid signature format",
			})
			return
		}

		if !hmac.Equal(expectedMAC, sigBytes) {
			fmt.Printf("Payload: %s\n", payload)
			fmt.Printf("Expected MAC: %s\n", expectedMAC)
			fmt.Printf("Received Signature: %s\n", signature)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid signature",
			})
			return
		}

		c.Next()
	}
}
