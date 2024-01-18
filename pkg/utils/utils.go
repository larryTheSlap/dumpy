package utils

import (
	"math/rand"
	"strconv"
)

func GenerateRandomID(digits int) (r string) {
	for i := 0; i < digits; i++ {
		r += strconv.Itoa(rand.Intn(10))
	}
	return
}

func ExistIn(t string, m map[string]string) (string, bool) {
	for l, s := range m {
		if t == l || t == s {
			return l, true
		}
	}
	return "", false
}
