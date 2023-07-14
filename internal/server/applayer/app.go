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
type Metrics storagelayer.Metrics
type Metric storagelayer.Metric
type app struct {
	store storagelayer.MemStorage
}

type App interface {
	GetAll() (map[string]float64, map[string]int64)
	Update(metricType, metricName, metricValue string) error
	Get(metricType, metricName string) (string, error)
	Ping() error
	Updates(metric Metrics) error
}

func New(store storagelayer.MemStorage) *app {
	return &app{store: store}
}

func checkIfWrongMetricType(metricType string) error {
	if metricType != "gauge" && metricType != "counter" {
		return ErrWrongMetricType
	}
	return nil
}

func (app *app) Get(metricType, metricName string) (string, error) {
	err := checkIfWrongMetricType(metricType)
	if err != nil {
		return "", err
	}
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

func (app *app) Update(metricType, metricName, metricValue string) error {
	err := checkIfWrongMetricType(metricType)
	if err != nil {
		return err
	}
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

func (app *app) Updates(metrics Metrics) error {
	for _, metric := range metrics {
		err := checkIfWrongMetricType(metric.MType)
		if err != nil {
			return err
		}
	}
	err := app.store.Updates(storagelayer.Metrics(metrics))
	if err != nil {
		return err
	}
	return nil
}

func (app *app) Ping() error {
	return app.store.CheckDBConnection()
}
