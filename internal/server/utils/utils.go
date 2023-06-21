package utils

import (
	"fmt"
	"strings"
)

func MetricValueWithoutTrailingZero(metricValue float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", metricValue), "0"), ".")
}
