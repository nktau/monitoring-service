package storagelayer

import (
	"errors"
)

var MetricNotFound = errors.New("metric not found in storage")

func (mem *memStorage) GetCounter(metricName string) (metricValue int64, err error) {
	value, ok := mem.counter[metricName]
	if ok {
		return value, nil
	}
	return -1, MetricNotFound
}

func (mem *memStorage) GetGauge(metricName string) (metricValue float64, err error) {
	value, ok := mem.gauge[metricName]
	if ok {
		return value, nil
	}
	return -1, MetricNotFound
}

func (mem *memStorage) GetAll() (map[string]float64, map[string]int64) {
	return mem.gauge, mem.counter
}
