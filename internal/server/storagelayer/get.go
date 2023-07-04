package storagelayer

import (
	"errors"
)

var ErrMetricNotFound = errors.New("metric not found")

func (mem *memStorage) GetCounter(metricName string) (metricValue int64, err error) {
	value, ok := mem.Counter[metricName]
	if ok {
		return value, nil
	}
	return -1, ErrMetricNotFound
}

func (mem *memStorage) GetGauge(metricName string) (metricValue float64, err error) {
	value, ok := mem.Gauge[metricName]
	if ok {
		return value, nil
	}
	return -1, ErrMetricNotFound
}

func (mem *memStorage) GetAll() (map[string]float64, map[string]int64) {
	return mem.Gauge, mem.Counter
}
