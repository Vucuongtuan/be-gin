package utils

import (
	"math/rand"
	"time"
)

func GenerateOtp() int {
	rand.Seed(time.Now().UnixNano())
	return 100000 + rand.Intn(900000)
}