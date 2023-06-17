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
	CommunicateWithHttpLayer(map[string]string) (string, error)
}

func New(store storagelayer.MemStorage) *app {
	return &app{store: store}
}

func (app *app) CommunicateWithHttpLayer(requestUrlMap map[string]string) (string, error) {
	if requestUrlMap["metricType"] != "gauge" && requestUrlMap["metricType"] != "counter" {
		return "", ErrWrongMetricType
	}
	if requestUrlMap["location"] == "update" {
		err := app.Update(requestUrlMap["metricType"], requestUrlMap["metricName"], requestUrlMap["metricValue"])
		if err != nil {
			return "", err
		}
	}
	if requestUrlMap["location"] == "value" {
		metricValue, err := app.Get(requestUrlMap["metricType"], requestUrlMap["metricName"])
		if err != nil {
			return "", err
		}
		return metricValue, nil
	}
	return "", nil
}
