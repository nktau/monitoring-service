package storagelayer

type memStorage struct {
	gauge   []map[string]float64
	counter []map[string]int64
}

type MemStorage interface {
	UpdateCounter(string, int64) error
	UpdateGauge(string, float64) error
	GetCounter(string) (int64, error)
	GetGauge(string) (float64, error)
	GetAll() ([]map[string]float64, []map[string]int64)
}

func New() *memStorage {
	return &memStorage{
		gauge:   []map[string]float64{},
		counter: []map[string]int64{},
	}
}
