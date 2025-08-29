package util

import (
	"math/rand"
	"time"
)

func GenerateRandomID() int64 {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	return time.Now().UnixNano() + int64(r.Intn(1000))
}
