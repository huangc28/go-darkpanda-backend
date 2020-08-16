package util

import (
	"math/rand"
	"time"
)

func Gen4DigitNum(low, high int) int {
	return low + rand.Intn(high-low)
}

var (
	charset               = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func GenRandStringRune(n int) string {
	pick := make([]byte, n)

	for idx := range pick {
		pick[idx] = charset[seededRand.Intn(len(charset))]
	}

	return string(pick)
}
