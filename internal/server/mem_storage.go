package server

type memStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() memStorage {
	var mem = memStorage{map[string]float64{}, map[string]int64{}}
	return mem
}

func (mem *memStorage) ModifyGauge(metricName string, metricValue float64) {
	mem.gauge[metricName] = metricValue
}

func (mem *memStorage) ModifyCounter(metricName string, metricValue int64) {
	mem.counter[metricName] += metricValue
}
