package applayer

func (app *app) UpdateCounter(s string, i int64) error {
	err := app.store.UpdateCounter(s, i)
	if err != nil {
		return err
	}
	return nil
}

func (app *app) UpdateGauge(s string, f float64) error {
	err := app.store.UpdateGauge(s, f)
	if err != nil {
		return err
	}
	return nil

}
