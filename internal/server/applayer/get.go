package applayer

func (app *app) GetCounter(metricName string) (int64, error) {
	metricValue, err := app.store.GetCounter(metricName)
	if err != nil {
		return -1, err
	}
	return metricValue, nil

}

func (app *app) GetGauge(metricName string) (float64, error) {
	metricValue, err := app.store.GetGauge(metricName)
	if err != nil {
		return -1, err
	}
	return metricValue, nil

}

func (app *app) GetAll() (map[string]float64, map[string]int64) {

	return app.store.GetAll()
}
