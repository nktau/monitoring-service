package storagelayer

import "fmt"

func (self *memStorage) UpdateGauge(metricName string, metricValue float64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("Error in storagelayer/UpdateGauge func %s", rec)
		}
	}()
	self.gauge[metricName] = metricValue
	return nil
}

func (self *memStorage) UpdateCounter(metricName string, metricValue int64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("Error in storagelayer/UpdateCounter func %s", rec)
		}
	}()
	self.counter[metricName] += metricValue
	return nil
}
