package storagelayer

type memStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

type MemStorage interface {
	UpdateCounter(string, int64) error
	UpdateGauge(string, float64) error
}

func New() *memStorage {
	var mem = &memStorage{map[string]float64{}, map[string]int64{}}
	return mem
}
