package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count     int
	resetTime time.Time
}

var visitors = make(map[string]*visitor)
var mu sync.Mutex

var maxRequests int

func init() {

	env := os.Getenv("RATE_LIMIT_PER_MINUTE")

	n, err := strconv.Atoi(env)
	if err != nil || n <= 0 {
		n = 3
	}

	maxRequests = n
}

func RateLimitByIP() gin.HandlerFunc {
	return func(c *gin.Context) {

		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]

		if !exists || time.Now().After(v.resetTime) {
			visitors[ip] = &visitor{
				count:     1,
				resetTime: time.Now().Add(time.Minute),
			}
			mu.Unlock()
			c.Next()
			return
		}

		if v.count >= maxRequests {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}

		v.count++
		mu.Unlock()

		c.Next()
	}
}
