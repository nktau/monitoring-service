package applayer

func (self *app) UpdateCounter(s string, i int64) error {
	err := self.store.UpdateCounter(s, i)
	if err != nil {
		return err
	}
	return nil
}

func (self *app) UpdateGauge(s string, f float64) error {
	err := self.store.UpdateGauge(s, f)
	if err != nil {
		return err
	}
	return nil

}
