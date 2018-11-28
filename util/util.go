package util

// Utilities related to random string generation and system time retrieving

import (
	"math/rand"
	"time"
)

// Initializes rand.Seed with current time
func init() {
	// time.Now().UnixNano() is the time in nanoseconds (since 01/01/1970 UTC)
	rand.Seed(time.Now().UnixNano())
}

var alphabet []rune = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_")

// Returns a random string with size equals @length
func RandomString(length int) string {
	res := make([]rune, length)
	for i := range res {
		res[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(res)
}

// Returns time in seconds (since 01/01/1970 UTC)
func Now() int64 {
	return time.Now().Unix()
}
