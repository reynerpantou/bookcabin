package randomutil

import "math/rand/v2"

func IsSuccess(rate *float64) bool {
	if rate == nil {
		return true
	}
	if *rate <= 0 {
		return false
	}
	if *rate >= 1 {
		return true
	}
	return rand.Float64() < *rate
}
