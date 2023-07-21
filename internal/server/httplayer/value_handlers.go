package httplayer

import (
	"encoding/json"
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

func (api *httpAPI) valueJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reader, err := api.readBody(r)
	if err != nil {
		api.logger.Error("can't decode request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(reader)
	r.Body.Close()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		api.logger.Info("can't read request body", zap.Error(err))
		return
	}
	api.logger.Debug("body:", zap.String("body", string(body)))
	var metric utils.Metrics
	err = json.Unmarshal(body, &metric)
	if err != nil {
		http.Error(w, "invalid json data", http.StatusBadRequest)
		api.logger.Info("get invalid json data from client",
			zap.String("data", string(body)),
			zap.Error(err),
		)
		return
	}
	// check if metric name is empty:
	if metric.ID == "" {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
		return
	}
	// check if metric type is empty:
	if metric.MType == "" {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
		return
	}
	value, errFromAppLayer := api.app.Get(metric.MType, metric.ID)
	err = handleApplayerValueError(errFromAppLayer, w)
	if err != nil {
		return
	}

	if metric.MType == "gauge" {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			api.logger.Error("can't convert from string to float64", zap.Error(err))
		}
		metric.Value = &floatValue
	} else {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			api.logger.Error("can't convert from string to float64", zap.Error(err))
		}
		metric.Delta = &intValue
	}
	responseBody, err := json.Marshal(metric)
	if err != nil {
		api.logger.Error("can't create json response body", zap.Error(err))
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (api *httpAPI) valuePlainText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	requestURLMap := map[string]string{}
	requestURLSlice := getRequestURLSlice(r.URL.Path)
	requestURLMap["location"] = requestURLSlice[0]
	// check if metric type is empty:
	if len(requestURLSlice) < 2 || requestURLSlice[1] == "" {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
		return
	}
	requestURLMap["metricType"] = requestURLSlice[1]
	// check if metric name is empty:
	if len(requestURLSlice) == 3 && requestURLSlice[2] == "" || len(requestURLSlice) == 2 {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
		return
	}
	requestURLMap["metricName"] = requestURLSlice[2]
	value, errFromAppLayer := api.app.Get(requestURLMap["metricType"], requestURLMap["metricName"])
	err := handleApplayerValueError(errFromAppLayer, w)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%s\n", value)))
}

func handleApplayerValueError(errFromAppLayer error, w http.ResponseWriter) error {
	if errFromAppLayer != nil {
		switch errFromAppLayer {
		case applayer.ErrWrongMetricType:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
			return errFromAppLayer
		case applayer.ErrWrongMetricName:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
			return errFromAppLayer
		case applayer.ErrMetricNotFound:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrMetricNotFound), http.StatusNotFound)
			return errFromAppLayer
		default:
			http.Error(w, fmt.Sprintf("%v", errFromAppLayer), http.StatusInternalServerError)
			return errFromAppLayer
		}
	}
	return nil
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(i))
	}
}
