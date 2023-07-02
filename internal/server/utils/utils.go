package utils

import (
	"strconv"
)

func MetricValueWithoutTrailingZero(metricValue float64) string {
	return strconv.FormatFloat(metricValue, 'f', -1, 64)
}
