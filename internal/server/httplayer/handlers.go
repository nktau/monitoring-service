package httplayer

import (
	"errors"
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"net/http"
	"strings"
)

var ErrMethodNotAllowed = errors.New("method not allowed")

func getRequestURLSlice(request string) []string {
	return strings.Split(strings.TrimLeft(request, "/"), "/")
}

func (api *httpAPI) update(w http.ResponseWriter, r *http.Request) {
	requestURLMap := map[string]string{}
	requestURLSlice := getRequestURLSlice(r.URL.Path)
	requestURLMap["location"] = requestURLSlice[0]
	// check if metric type is empty:
	if len(requestURLSlice) < 2 || requestURLSlice[1] == "" {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
		return
	}
	requestURLMap["metricType"] = requestURLSlice[1]
	//
	// check if metric name is empty:
	if len(requestURLSlice) == 3 && requestURLSlice[2] == "" || len(requestURLSlice) == 2 {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
		return
	}
	requestURLMap["metricName"] = requestURLSlice[2]
	// check if metric value is empty:
	if len(requestURLSlice) < 4 {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricValue), http.StatusBadRequest)
		return
	}
	if len(requestURLSlice) >= 4 {
		requestURLMap["metricValue"] = requestURLSlice[3]
	}
	_, err := api.app.ParseUpdateAndValue(requestURLMap)
	if err != nil {
		switch err {
		case applayer.ErrWrongMetricType:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
			return
		case applayer.ErrWrongMetricName:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
			return
		case applayer.ErrWrongMetricValue:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricValue), http.StatusBadRequest)
			return
		default:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("ok\n"))
}

func (api *httpAPI) value(w http.ResponseWriter, r *http.Request) {
	requestURLMap := map[string]string{}
	requestURLSlice := getRequestURLSlice(r.URL.Path)
	requestURLMap["location"] = requestURLSlice[0]
	// check if metric type is empty:
	if len(requestURLSlice) < 2 || requestURLSlice[1] == "" {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
		return
	}
	requestURLMap["metricType"] = requestURLSlice[1]
	//
	// check if metric name is empty:
	if len(requestURLSlice) == 3 && requestURLSlice[2] == "" || len(requestURLSlice) == 2 {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
		return
	}
	requestURLMap["metricName"] = requestURLSlice[2]
	value, err := api.app.ParseUpdateAndValue(requestURLMap)
	if err != nil {
		switch err {
		case applayer.ErrWrongMetricType:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
			return
		case applayer.ErrWrongMetricName:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
			return
		case applayer.ErrMetricNotFound:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrMetricNotFound), http.StatusNotFound)
			return
		default:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte(fmt.Sprintf("%s\n", value)))
}

func (api *httpAPI) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	gauge, counter := api.app.GetAll()
	var s []string
	for key, value := range gauge {
		s = append(s, fmt.Sprintf("<h3>%s: %s</h3>\n", key, utils.MetricValueWithoutTrailingZero(value)))
	}
	for key, value := range counter {
		s = append(s, fmt.Sprintf("<h3>%s: %d</h3>\n", key, value))
	}
	for _, i := range s {
		w.Write([]byte(i))
	}
}
