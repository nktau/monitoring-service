package storagelayer

import (
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/config"
)

func (mem *memStorage) UpdateGauge(metricName string, metricValue float64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateGauge func %s", rec)
		}
	}()
	mem.Gauge[metricName] = metricValue
	if config.Config.StoreInterval == 0 && config.Config.FileStoragePath != "" {
		err := mem.writeToDisk()
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *memStorage) UpdateCounter(metricName string, metricValue int64) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error in storagelayer/UpdateCounter func %s", rec)
		}
	}()
	mem.Counter[metricName] += metricValue
	if config.Config.StoreInterval == 0 && config.Config.FileStoragePath != "" {
		err := mem.writeToDisk()
		if err != nil {
			return err
		}
	}
	return nil
}
