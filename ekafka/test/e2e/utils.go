package e2e

import (
	"math/rand"
	"time"
)

const charPool = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = charPool[rand.Intn(len(charPool))]
	}
	return string(b)
}
