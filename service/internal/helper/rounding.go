package helper

import "math"

func DecimalPlaces(rounding float64) int32 {
	if rounding <= 0 {
		return 2
	}
	dp := int32(math.Round(math.Log10(1 / rounding)))
	if dp < 0 {
		return 0
	}
	return dp
}
