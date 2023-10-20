package number

import "math"

// Round rounds the provided value to nearest number.
func Round(value float64, decimals int) float64 {
	multiplier := float64(1.0)
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	return math.Round(value*multiplier) / multiplier
}
