package genverifycode

import (
	"fmt"
	"math/rand"
	"time"
)

var (
	charset               = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func GenNum(low, high int) int {
	return low + seededRand.Intn(high-low)
}

func GenRandStringRune(n int) string {
	pick := make([]byte, n)

	for idx := range pick {
		pick[idx] = charset[seededRand.Intn(len(charset))]
	}

	return string(pick)
}

type VerifyCode struct {
	Dig   int
	Chars string
}

func (vs *VerifyCode) BuildCode() string {
	return fmt.Sprintf("%s-%d", vs.Chars, vs.Dig)
}

func GenVerifyCode() VerifyCode {
	return VerifyCode{
		Dig:   GenNum(1000, 9999),
		Chars: GenRandStringRune(3),
	}
}
