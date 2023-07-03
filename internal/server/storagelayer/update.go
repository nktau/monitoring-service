package storagelayer

import "fmt"

func (mem *memStorage) UpdateGauge(metricName string, metricValue float64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateGauge func %s", rec)
		}
	}()
	mem.gauge = append(mem.gauge, map[string]float64{metricName: metricValue})
	return nil
}

func (mem *memStorage) UpdateCounter(metricName string, metricValue int64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateCounter func %s", rec)
		}
	}()
	isMetricNameInStore := false
	for _, value := range mem.counter {
		if _, ok := value[metricName]; ok {
			value[metricName] += metricValue
			isMetricNameInStore = true
			return nil
		}
	}
	if !isMetricNameInStore {
		mem.counter = append(mem.counter, map[string]int64{metricName: metricValue})
	}

	return nil
}
