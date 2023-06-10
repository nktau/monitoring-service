package httplayer

import (
	"fmt"
	"github.com/nktau/monitoring-service/internal/applayer"
	"net/http"
	"net/url"
	"strings"
)

func (api *httpAPI) value(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	u, err := url.Parse(strings.TrimPrefix(r.URL.Path, "/value/"))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	pathSlice := strings.Split(u.Path, "/")
	if pathSlice[0] != "gauge" && pathSlice[0] != "counter" {
		http.Error(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if len(pathSlice) == 2 && pathSlice[1] == "" || len(pathSlice) == 1 {
		http.Error(w, "wrong metric name", http.StatusNotFound)
		return
	}

	if pathSlice[0] == "gauge" {
		metricValue, err := api.app.GetGauge(pathSlice[1])
		if err == applayer.ErrMetricNotFound {
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		metricValueWithoutTrailingZero := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", metricValue), "0"), ".")

		w.Write([]byte(metricValueWithoutTrailingZero + "\n"))
		return
	}
	if pathSlice[0] == "counter" {
		metricValue, err := api.app.GetCounter(pathSlice[1])
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
