package applayer

import (
	"errors"
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/storagelayer"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"strconv"
)

var ErrMetricNotFound = storagelayer.ErrMetricNotFound     // http.StatusNotFound
var ErrWrongMetricType = errors.New("wrong metric type")   // http.StatusBadRequest
var ErrWrongMetricName = errors.New("wrong metric name")   // http.StatusNotFound
var ErrWrongMetricValue = errors.New("wrong metric value") // http.StatusBadRequest

type app struct {
	store storagelayer.MemStorage
}

type App interface {
	GetAll() (map[string]float64, map[string]int64)
	ParseUpdateAndValue(map[string]string) (string, error)
}

func New(store storagelayer.MemStorage) *app {
	return &app{store: store}
}

func (app *app) ParseUpdateAndValue(requestURLMap map[string]string) (string, error) {
	if requestURLMap["metricType"] != "gauge" && requestURLMap["metricType"] != "counter" {
		return "", ErrWrongMetricType
	}
	if requestURLMap["location"] == "update" {
		err := app.update(requestURLMap["metricType"], requestURLMap["metricName"], requestURLMap["metricValue"])
		if err != nil {
			return "", err
		}
	}
	if requestURLMap["location"] == "value" {
		metricValue, err := app.get(requestURLMap["metricType"], requestURLMap["metricName"])
		if err != nil {
			return "", err
		}
		return metricValue, nil
	}
	return "", nil
}

func (app *app) get(metricType, metricName string) (string, error) {
	if metricType == "gauge" {
		metricValue, err := app.store.GetGauge(metricName)
		if err != nil {
			return "", err
		}
		return utils.MetricValueWithoutTrailingZero(metricValue), nil
	}
	if metricType == "counter" {
		metricValue, err := app.store.GetCounter(metricName)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(metricValue, 10), nil
	}
	return "", fmt.Errorf("uncatched error")

}

func (app *app) GetAll() (map[string]float64, map[string]int64) {

	return app.store.GetAll()
}

func (app *app) update(metricType, metricName, metricValue string) error {
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
