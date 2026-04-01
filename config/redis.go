package config

import (
	"context"
	"crypto/tls"
	"go-api/helpers"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()
var Redis *redis.Client

func InitRedis() {

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	username := os.Getenv("REDIS_USERNAME")

	dbStr := os.Getenv("REDIS_DB")
	if dbStr == "" {
		dbStr = "0"
	}

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		log.Fatal("Invalid REDIS_DB value")
	}

	addr := host + ":" + port

	useTLS := os.Getenv("REDIS_TLS") == "true"

	opt := &redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	}

	if useTLS {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	Redis = redis.NewClient(opt)

	// Test connection
	_, err = Redis.Ping(Ctx).Result()
	if err != nil {
		helpers.ErrorLogger.Fatal("Failed to connect to Redis:", err)
	}

	helpers.InfoLogger.Println("Connected to Redis successfully")
}
