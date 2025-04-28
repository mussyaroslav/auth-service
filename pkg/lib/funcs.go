package lib

import (
	"math"
)

// MaskedText возвращает замаскированную точками строку, оставив слева и справа оригинал в n символов
func MaskedText(msg string, n int) string {
	c := int(math.Floor(float64(len(msg)) / 3))
	if c > n {
		c = n
	}
	if c <= 0 {
		c = 1
	}
	return msg[0:c] + "..." + msg[len(msg)-c:]
}
