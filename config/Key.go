package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var JWTKeyString string
var HMACSecret string
var JWTKey []byte

func LoadEnv() {
	env := godotenv.Load()
	if env != nil {
		panic("Error loading .env file")
	}
	JWTKeyString = os.Getenv("JWT_KEY")
	HMACSecret = os.Getenv("HMAC_SECRET")
	if JWTKeyString == "" {
		log.Println("WARNING: JWT_KEY not set, using default")
		JWTKeyString = "mysupersecretkeymustbe32bytes!!!"
	}

	if HMACSecret == "" {
		log.Println("WARNING: HMAC_SECRET not set, using default")
		HMACSecret = "mysupersecretkeymustbe32bytes!!!"
	}

	JWTKey = []byte(JWTKeyString)
}
