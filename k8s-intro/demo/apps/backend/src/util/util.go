package util

import (
	"math/rand"
	"time"
)

// HandleError checks if an error exists and panics if it does
func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

// RandomString generates a random string of the given length
func RandomString(length int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
