package utils

import (
	"math"
	"time"
)

func ExponentialBackoff(retries int, baseDuration time.Duration) time.Duration {
	if retries <= 0 {
		return 0
	}
	return time.Duration(math.Pow(2, float64(retries-1))) * baseDuration
}
