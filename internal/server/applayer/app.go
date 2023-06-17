package applayer

import (
	"errors"
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/storagelayer"
	"strings"
)

var ErrMetricNotFound = storagelayer.ErrMetricNotFound
var ErrWrongMetricType = errors.New("wrong metric type")   // http.StatusBadRequest
var ErrWrongMetricName = errors.New("wrong metric name")   // http.StatusNotFound
var ErrWrongMetricValue = errors.New("wrong metric value") // http.StatusBadRequest

func metricValueWithoutTrailingZero(metricValue float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", metricValue), "0"), ".")
}

type app struct {
	store storagelayer.MemStorage
}

type App interface {
	Update(string, string, string) error
	Get(string, string) (string, error)
	GetAll() (map[string]float64, map[string]int64)
	CommunicateWithHTTPLayer(map[string]string) (string, error)
}

func New(store storagelayer.MemStorage) *app {
	return &app{store: store}
}

func (app *app) CommunicateWithHTTPLayer(requestURLMap map[string]string) (string, error) {
	if requestURLMap["metricType"] != "gauge" && requestURLMap["metricType"] != "counter" {
		return "", ErrWrongMetricType
	}
	if requestURLMap["location"] == "update" {
		err := app.Update(requestURLMap["metricType"], requestURLMap["metricName"], requestURLMap["metricValue"])
		if err != nil {
			return "", err
		}
	}
	if requestURLMap["location"] == "value" {
		metricValue, err := app.Get(requestURLMap["metricType"], requestURLMap["metricName"])
		if err != nil {
			return "", err
		}
		return metricValue, nil
	}
	return "", nil
}
