package helpers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func GenerateOTP() (string, string) {
	otp := rand.New(rand.NewSource(time.Now().UnixNano()))
	challengeID := uuid.NewString()
	return challengeID, fmt.Sprintf("%06d", otp.Intn(1000000))
}
