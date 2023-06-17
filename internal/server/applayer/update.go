package applayer

import (
	"strconv"
)

func (app *app) Update(metricType, metricName, metricValue string) error {
	if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ErrWrongMetricValue
		}
		err = app.store.UpdateGauge(metricName, value)
		if err != nil {
			return err
		}
	}
	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return ErrWrongMetricValue
		}
		err = app.store.UpdateCounter(metricName, value)
		if err != nil {
			return err
		}
	}
	return nil
}
