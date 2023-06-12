package httplayer

import (
	"fmt"
	"github.com/nktau/monitoring-service/internal/applayer"
	"net/http"
	"strconv"
	"strings"
)

func metricValueWithoutTrailingZero(metricValue float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", metricValue), "0"), ".")
}

func (api *httpAPI) update(w http.ResponseWriter, r *http.Request) {

	if r.Context().Value(contextKeyMetricType) == "gauge" {
		value, err := strconv.ParseFloat(r.Context().Value(contextKeyMetricValue).(string), 64)
		if err != nil {
			fmt.Println(r.Context().Value(contextKeyMetricValue).(string))
			http.Error(w, "wrong metric value", http.StatusBadRequest)
			return
		}
		api.app.UpdateGauge(r.Context().Value(contextKeyMetricName).(string), value)
	}
	if r.Context().Value(contextKeyMetricType) == "counter" {
		value, err := strconv.ParseInt(r.Context().Value(contextKeyMetricValue).(string), 10, 64)
		if err != nil {
			http.Error(w, "wrong metric value", http.StatusBadRequest)
			return
		}
		api.app.UpdateCounter(r.Context().Value(contextKeyMetricName).(string), value)
	}

	fmt.Println(r.Context().Value(contextKeyMetricType))
	w.Write([]byte("ok\n"))
}

func (api *httpAPI) value(w http.ResponseWriter, r *http.Request) {
	if r.Context().Value(contextKeyMetricType) == "gauge" {
		metricValue, err := api.app.GetGauge(r.Context().Value(contextKeyMetricName).(string))
		if err == applayer.ErrMetricNotFound {
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(metricValueWithoutTrailingZero(metricValue) + "\n"))
		return
	}
	if r.Context().Value(contextKeyMetricType) == "counter" {
		metricValue, err := api.app.GetCounter(r.Context().Value(contextKeyMetricName).(string))
		if err == applayer.ErrMetricNotFound {
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		s := fmt.Sprintf("%d\n", metricValue)
		w.Write([]byte(s))
		return
	}
}

func (api *httpAPI) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	gauge, counter := api.app.GetAll()
	var s []string
	for key, value := range gauge {
		//to do: create a function
		s = append(s, fmt.Sprintf("<h3>%s: %s</h3>\n", key, metricValueWithoutTrailingZero(value)))
	}
	for key, value := range counter {
		s = append(s, fmt.Sprintf("<h3>%s: %d</h3>\n", key, value))
	}
	for _, i := range s {
		w.Write([]byte(i))
	}
}
