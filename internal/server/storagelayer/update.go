package storagelayer

import "fmt"

func (mem *memStorage) UpdateGauge(metricName string, metricValue float64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateGauge func %s", rec)
		}
	}()
	mem.gauge[metricName] = metricValue
	return nil
}

func (mem *memStorage) UpdateCounter(metricName string, metricValue int64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateCounter func %s", rec)
		}
	}()
	mem.counter[metricName] += metricValue
	return nil
}
