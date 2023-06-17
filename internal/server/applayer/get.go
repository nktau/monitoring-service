package applayer

import "fmt"

func (app *app) Get(metricType, metricName string) (string, error) {
	if metricType == "gauge" {
		metricValue, err := app.store.GetGauge(metricName)
		if err != nil {
			return "", err
		}
		return metricValueWithoutTrailingZero(metricValue), nil
	}
	if metricType == "counter" {
		metricValue, err := app.store.GetCounter(metricName)
		if err != nil {
			return "", err
		}
		return string(metricValue), nil
	}
	return "", fmt.Errorf("uncatched error")

}

func (app *app) GetAll() (map[string]float64, map[string]int64) {

	return app.store.GetAll()
}
