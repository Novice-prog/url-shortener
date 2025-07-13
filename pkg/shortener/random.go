package shortener

import (
	"crypto/rand"
	"math/big"
)

const urlSafeChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomURL(length int) string {
	if length <= 0 {
		length = 8
	}

	// Максимальное значение для rand.Int
	max := big.NewInt(int64(len(urlSafeChars)))
	result := make([]byte, length)

	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// На редкий случай ошибки rand: deterministic fallback
			result[i] = urlSafeChars[i%len(urlSafeChars)]
			continue
		}
		result[i] = urlSafeChars[n.Int64()]
	}

	return string(result)
}
